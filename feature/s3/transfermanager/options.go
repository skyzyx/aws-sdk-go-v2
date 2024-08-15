package transfermanager

// Options provides params needed for transfer api calls
type Options struct {
	// The client to use when uploading to S3.
	S3 S3Client

	// The buffer size (in bytes) to use when buffering data into chunks and
	// sending them as parts to S3. The minimum allowed part size is 5MB, and
	// if this value is set to zero, the DefaultUploadPartSize value will be used.
	PartSizeBytes int64

	// The threshold bytes to decide when the file should be multi-uploaded
	MultipartUploadThreshold int64

	// Option to disable checksum validation for download
	DisableChecksum bool

	// Checksum algorithm to use for upload
	ChecksumAlgorithm ChecksumAlgorithm

	// The number of goroutines to spin up in parallel per call to Upload when
	// sending parts. If this is set to zero, the DefaultUploadConcurrency value
	// will be used.
	//
	// The concurrency pool is not shared between calls to Upload.
	Concurrency int
}

func (o *Options) init() {
	if o.Concurrency == 0 {
		o.Concurrency = DefaultTransferConcurrency
	}
	if o.PartSizeBytes == 0 {
		o.PartSizeBytes = DefaultPartSizeBytes
	}
	if o.ChecksumAlgorithm == "" {
		o.ChecksumAlgorithm = ChecksumAlgorithmCrc32
	}
	if o.MultipartUploadThreshold == 0 {
		o.MultipartUploadThreshold = DefaultMultipartUploadThreshold
	}
}

// Copy returns new copy of the Options
func (o Options) Copy() Options {
	to := o
	return to
}
