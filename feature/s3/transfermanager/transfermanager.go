package transfermanager

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

const userAgentKey = "s3-transfer"

// MaxUploadParts is the maximum allowed number of parts in a multi-part upload
// on Amazon S3.
const DefaultMaxUploadParts int64 = 10000

// DefaultPartSizeBytes is the default part size when transferring objects to/from S3
const DefaultPartSizeBytes int64 = 1024 * 1024 * 8

const DefaultMPUThreshold int64 = 1024 * 1024 * 16

// DefaultTransferConcurrency is the default number of goroutines to spin up when
// using PutObject().
const DefaultTransferConcurrency = 5

type Client struct {
	options Options
}

// New returns an initialized Client from the client Options. Provide
// more functional options to further configure the Client
func New(opts Options, optFns ...func(*Options)) *Client {
	for _, fn := range optFns {
		fn(&opts)
	}

	return &Client{
		options: opts,
	}
}

// NewFromConfig returns a new Client from the provided s3 config
func NewFromConfig(cfg aws.Config, optFns ...func(*s3.Options)) *Client {
	return New(Options{
		S3: s3.NewFromConfig(cfg, optFns...),
	})
}

type Options struct {
	// The client to use when uploading to S3.
	S3 S3APIClient

	// The buffer size (in bytes) to use when buffering data into chunks and
	// sending them as parts to S3. The minimum allowed part size is 5MB, and
	// if this value is set to zero, the DefaultUploadPartSize value will be used.
	PartSizeBytes int64

	// the threshold bytes to decide when the file should be multi-uploaded
	MPUThreshold int64

	DisableChecksum bool

	ChecksumAlgorithm types.ChecksumAlgorithm

	MaxUploadParts int64

	// The number of goroutines to spin up in parallel per call to Upload when
	// sending parts. If this is set to zero, the DefaultUploadConcurrency value
	// will be used.
	//
	// The concurrency pool is not shared between calls to Upload.
	Concurrency int

	// List of request options that will be passed down to individual PutObject
	// operation requests
	PutClientOptions []func(*s3.Options)

	// partPool allows for the re-usage of streaming payload part buffers between upload calls
	partPool byteSlicePool
}

func (o *Options) init() error {
	if o.S3 == nil {
		cfg, err := config.LoadDefaultConfig(context.TODO())
		if err != nil {
			return fmt.Errorf("error while creating default s3 cfg: %q", err)
		}
		o.S3 = s3.NewFromConfig(cfg)
	}

	if o.Concurrency == 0 {
		o.Concurrency = DefaultTransferConcurrency
	}
	if o.PartSizeBytes == 0 {
		o.PartSizeBytes = DefaultPartSizeBytes
	} else if o.PartSizeBytes < DefaultPartSizeBytes {
		return fmt.Errorf("part size must be at least %d bytes", DefaultPartSizeBytes)
	}
	if o.MaxUploadParts == 0 {
		o.MaxUploadParts = DefaultMaxUploadParts
	}
	if o.ChecksumAlgorithm == "" {
		o.ChecksumAlgorithm = types.ChecksumAlgorithmCrc32
	}

	// If PartSize was changed or partPool was never setup then we need to allocated a new pool
	// so that we return []byte slices of the correct size
	poolCap := o.Concurrency + 1
	if o.partPool == nil || o.partPool.SliceSize() != o.PartSizeBytes {
		o.partPool = newByteSlicePool(o.PartSizeBytes)
		o.partPool.ModifyCapacity(poolCap)
	} else {
		o.partPool = &returnCapacityPoolCloser{byteSlicePool: o.partPool}
		o.partPool.ModifyCapacity(poolCap)
	}

	return nil
}
