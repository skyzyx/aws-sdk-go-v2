package transfermanager

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/internal/awsutil"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	smithymiddleware "github.com/aws/smithy-go/middleware"
)

// A MultiUploadFailure wraps a failed S3 multipart upload. An error returned
// will satisfy this interface when a multi part upload failed to upload all
// chucks to S3. In the case of a failure the UploadID is needed to operate on
// the chunks, if any, which were uploaded.
//
// Example:
//
//	u := manager.NewUploader(client)
//	output, err := u.upload(context.Background(), input)
//	if err != nil {
//		var multierr manager.MultiUploadFailure
//		if errors.As(err, &multierr) {
//			fmt.Printf("upload failure UploadID=%s, %s\n", multierr.UploadID(), multierr.Error())
//		} else {
//			fmt.Printf("upload failure, %s\n", err.Error())
//		}
//	}
type MultiUploadFailure interface {
	error

	// UploadID returns the upload id for the S3 multipart upload that failed.
	UploadID() string
}

// A multiUploadError wraps the upload ID of a failed s3 multipart upload.
// Composed of BaseError for code, message, and original error
//
// Should be used for an error that occurred failing a S3 multipart upload,
// and a upload ID is available. If an uploadID is not available a more relevant
type multiUploadError struct {
	err error

	// ID for multipart upload which failed.
	uploadID string
}

// batchItemError returns the string representation of the error.
//
// # See apierr.BaseError ErrorWithExtra for output format
//
// Satisfies the error interface.
func (m *multiUploadError) Error() string {
	var extra string
	if m.err != nil {
		extra = fmt.Sprintf(", cause: %s", m.err.Error())
	}
	return fmt.Sprintf("upload multipart failed, upload id: %s%s", m.uploadID, extra)
}

// Unwrap returns the underlying error that cause the upload failure
func (m *multiUploadError) Unwrap() error {
	return m.err
}

// UploadID returns the id of the S3 upload which failed.
func (m *multiUploadError) UploadID() string {
	return m.uploadID
}

type PutObjectInput struct {
	// Bucket the object is uploaded into
	Bucket *string

	// Object key for which the PUT action was initiated
	Key *string

	// Object data
	Body io.Reader

	// Indicates the algorithm used to create the checksum for the object
	ChecksumAlgorithm ChecksumAlgorithm
}

// PutObjectOutput represents a response from the PutObject() call.
type PutObjectOutput struct {
	// The URL where the object was uploaded to.
	Location string

	// The ID for a multipart upload to S3. In the case of an error the error
	// can be cast to the MultiUploadFailure interface to extract the upload ID.
	// Will be empty string if multipart upload was not used, and the object
	// was uploaded as a single PutObject call.
	UploadID string

	// The list of parts that were uploaded and their checksums. Will be empty
	// if multipart upload was not used, and the object was uploaded as a
	// single PutObject call.
	CompletedParts []types.CompletedPart

	// Indicates whether the uploaded object uses an S3 Bucket Key for server-side
	// encryption with Amazon Web Services KMS (SSE-KMS).
	BucketKeyEnabled bool

	// The base64-encoded, 32-bit CRC32 checksum of the object.
	ChecksumCRC32 *string

	// The base64-encoded, 32-bit CRC32C checksum of the object.
	ChecksumCRC32C *string

	// The base64-encoded, 160-bit SHA-1 digest of the object.
	ChecksumSHA1 *string

	// The base64-encoded, 256-bit SHA-256 digest of the object.
	ChecksumSHA256 *string

	// Entity tag for the uploaded object.
	ETag *string

	// If the object expiration is configured, this will contain the expiration date
	// (expiry-date) and rule ID (rule-id). The value of rule-id is URL encoded.
	Expiration *string

	// The object key of the newly created object.
	Key *string

	// If present, indicates that the requester was successfully charged for the
	// request.
	RequestCharged types.RequestCharged

	// If present, specifies the ID of the Amazon Web Services Key Management Service
	// (Amazon Web Services KMS) symmetric customer managed customer master key (CMK)
	// that was used for the object.
	SSEKMSKeyId *string

	// If you specified server-side encryption either with an Amazon S3-managed
	// encryption key or an Amazon Web Services KMS customer master key (CMK) in your
	// initiate multipart upload request, the response includes this header. It
	// confirms the encryption algorithm that Amazon S3 used to encrypt the object.
	ServerSideEncryption types.ServerSideEncryption

	// The version of the object that was uploaded. Will only be populated if
	// the S3 Bucket is versioned. If the bucket is not versioned this field
	// will not be set.
	VersionID *string
}

func (c *Client) Upload(ctx context.Context, input *PutObjectInput, opts ...func(*Options)) (*PutObjectOutput, error) {
	// Copy ClientOptions
	clientOptions := make([]func(*s3.Options), 0, len(c.options.PutClientOptions)+1)
	clientOptions = append(clientOptions, func(o *s3.Options) {
		o.APIOptions = append(o.APIOptions,
			middleware.AddSDKAgentKey(middleware.FeatureMetadata, userAgentKey),
			addFeatureUserAgent, // yes, there are two of these
			func(s *smithymiddleware.Stack) error {
				return s.Finalize.Insert(&setS3ExpressDefaultChecksum{}, "ResolveEndpointV2", smithymiddleware.After)
			},
		)
	})
	clientOptions = append(clientOptions, c.options.PutClientOptions...)
	c.options.PutClientOptions = clientOptions
	for _, opt := range opts {
		opt(&c.options)
	}

	i := uploader{in: input, cfg: c, ctx: ctx}

	return i.upload()
}

type uploader struct {
	ctx context.Context
	cfg *Client

	in *PutObjectInput

	readerPos int64 // current reader position
	totalSize int64 // set to -1 if the size is not known
}

func (u *uploader) upload() (*PutObjectOutput, error) {
	if err := u.cfg.options.init(); err != nil {
		return nil, fmt.Errorf("unable to initialize upload: %w", err)
	}
	defer u.cfg.options.partPool.Close()

	r, _, cleanUp, err := u.nextReader()

	if err == io.EOF {
		return u.singleUpload(r, cleanUp)
	} else if err != nil {
		cleanUp()
		return nil, err
	}

	mu := multiUploader{
		uploader: u,
	}
	return mu.upload(r, cleanUp)
}

func (u *uploader) singleUpload(r io.Reader, cleanUp func()) (*PutObjectOutput, error) {
	defer cleanUp()

	var params s3.PutObjectInput
	awsutil.Copy(params, u.in)

	var locationRecorder wrappedLocationClient
	resp, err := u.cfg.options.S3.PutObject(u.ctx, &params, append(u.cfg.options.PutClientOptions, locationRecorder.wrapClient())...)
	if err != nil {
		return nil, err
	}

	return &PutObjectOutput{}, nil
}

type httpClient interface {
	Do(r *http.Request) (*http.Response, error)
}

type wrappedLocationClient struct {
	httpClient
	location string
}

func (c *wrappedLocationClient) wrapClient() func(*s3.Options) {
	return func(o *s3.Options) {
		c.httpClient = o.HTTPClient
		o.HTTPClient = c
	}
}

func (c *wrappedLocationClient) Do(r *http.Request) (resp *http.Response, err error) {
	resp, err = c.httpClient.Do(r)
	if err != nil {
		return resp, err
	}

	if resp.Request != nil && resp.Request.URL != nil {
		url = *resp.Request.URL
		url.RawQuery = ""
		c.location = url.String()
	}
	return resp, err
}

// nextReader reads the next chunk of data from input Body
func (u *uploader) nextReader() (io.Reader, int, func(), error) {
	part, err := u.cfg.options.partPool.Get(u.ctx)
	if err != nil {
		return nil, 0, func() {}, err
	}

	n, err := readFillBuf(u.in.Body, *part)

	u.readerPos += int64(n)

	cleanUp := func() {
		u.cfg.options.partPool.Put(part)
	}
	return bytes.NewReader((*part)[0:n]), n, cleanUp, err
}

func readFillBuf(r io.Reader, b []byte) (offset int, err error) {
	for offset < len(b) && err == nil {
		var n int
		n, err = r.Read(b[offset:])
		offset += n
	}
	return offset, err
}

type multiUploader struct {
	*uploader
	wg       sync.WaitGroup
	m        sync.Mutex
	err      error
	uploadID string
	parts    completedParts
}

type completedParts []types.CompletedPart

func (cp completedParts) Len() int {
	return len(cp)
}

func (cp completedParts) Less(i, j int) bool {
	return aws.ToInt32(cp[i].PartNumber) < aws.ToInt32(cp[j].PartNumber)
}

func (cp completedParts) Swap(i, j int) {
	cp[i], cp[j] = cp[j], cp[i]
}

// upload will perform a multipart upload using the firstBuf buffer containing
// the first chunk of data.
func (u *multiUploader) upload(firstBuf io.Reader, cleanup func()) (*PutObjectOutput, error) {
	var params s3.CreateMultipartUploadInput
	awsutil.Copy(&params, u.uploader.in)

}

// setS3ExpressDefaultChecksum defaults to CRC32 for S3Express buckets,
// which is required when uploading to those through transfer manager.
type setS3ExpressDefaultChecksum struct{}

func (*setS3ExpressDefaultChecksum) ID() string {
	return "setS3ExpressDefaultChecksum"
}

func (*setS3ExpressDefaultChecksum) HandleFinalize(
	ctx context.Context, in smithymiddleware.FinalizeInput, next smithymiddleware.FinalizeHandler,
) (
	out smithymiddleware.FinalizeOutput, metadata smithymiddleware.Metadata, err error,
) {
	const checksumHeader = "x-amz-checksum-algorithm"

	if internalcontext.GetS3Backend(ctx) != internalcontext.S3BackendS3Express {
		return next.HandleFinalize(ctx, in)
	}

	// If this is CreateMultipartUpload we need to ensure the checksum
	// algorithm header is present. Otherwise everything is driven off the
	// context setting and we can let it flow from there.
	if middleware.GetOperationName(ctx) == "CreateMultipartUpload" {
		r, ok := in.Request.(*smithyhttp.Request)
		if !ok {
			return out, metadata, fmt.Errorf("unknown transport type %T", in.Request)
		}

		if internalcontext.GetChecksumInputAlgorithm(ctx) == "" {
			r.Header.Set(checksumHeader, "CRC32")
		}
		return next.HandleFinalize(ctx, in)
	} else if internalcontext.GetChecksumInputAlgorithm(ctx) == "" {
		ctx = internalcontext.SetChecksumInputAlgorithm(ctx, string(types.ChecksumAlgorithmCrc32))
	}

	return next.HandleFinalize(ctx, in)
}

func addFeatureUserAgent(stack *smithymiddleware.Stack) error {
	ua, err := getOrAddRequestUserAgent(stack)
	if err != nil {
		return err
	}

	ua.AddUserAgentFeature(middleware.UserAgentFeatureS3Transfer)
	return nil
}

func getOrAddRequestUserAgent(stack *smithymiddleware.Stack) (*middleware.RequestUserAgent, error) {
	id := (*middleware.RequestUserAgent)(nil).ID()
	mw, ok := stack.Build.Get(id)
	if !ok {
		mw = middleware.NewRequestUserAgent()
		if err := stack.Build.Add(mw, smithymiddleware.After); err != nil {
			return nil, err
		}
	}

	ua, ok := mw.(*middleware.RequestUserAgent)
	if !ok {
		return nil, fmt.Errorf("%T for %s middleware did not match expected type", mw, id)
	}

	return ua, nil
}
