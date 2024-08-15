package transfermanager

import (
	"github.com/aws/aws-sdk-go-v2/aws"
)

const userAgentKey = "s3-transfer"

// DefaultMaxUploadParts is the maximum allowed number of parts in a multi-part upload
// on Amazon S3.
const DefaultMaxUploadParts int32 = 10000

// DefaultPartSizeBytes is the default part size when transferring objects to/from S3
const DefaultPartSizeBytes int64 = 1024 * 1024 * 8

// DefaultMPUThreshold is the default size threshold in bytes indicating when to use multipart upload.
const DefaultMultipartUploadThreshold int64 = 1024 * 1024 * 16

// DefaultTransferConcurrency is the default number of goroutines to spin up when
// using PutObject().
const DefaultTransferConcurrency = 5

// Client provides the API client to make operations call for Amazon Simple
// Storage Service's Transfer Manager
type Client struct {
	options Options
}

// New returns an initialized Client from the client Options. Provide
// more functional options to further configure the Client
func New(s3Client S3Client, opts Options, optFns ...func(*Options)) *Client {
	opts.S3 = s3Client
	for _, fn := range optFns {
		fn(&opts)
	}

	opts.init()

	return &Client{
		options: opts,
	}
}

// NewFromConfig returns a new Client from the provided s3 config
func NewFromConfig(s3Client S3Client, cfg aws.Config, optFns ...func(*Options)) *Client {
	return New(s3Client, Options{}, optFns...)
}
