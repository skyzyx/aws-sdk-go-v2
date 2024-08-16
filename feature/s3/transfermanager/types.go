package transfermanager

import (
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// ChecksumAlgorithm indicates the algorithm used to create the checksum for the object
type ChecksumAlgorithm string

// Enum values for ChecksumAlgorithm
const (
	ChecksumAlgorithmCrc32  ChecksumAlgorithm = "CRC32"
	ChecksumAlgorithmCrc32c ChecksumAlgorithm = "CRC32C"
	ChecksumAlgorithmSha1   ChecksumAlgorithm = "SHA1"
	ChecksumAlgorithmSha256 ChecksumAlgorithm = "SHA256"
)

// ObjectCannedACL defines the canned ACL to apply to the object, see [Canned ACL] in the
// Amazon S3 User Guide.
type ObjectCannedACL string

// Enum values for ObjectCannedACL
const (
	ObjectCannedACLPrivate                ObjectCannedACL = "private"
	ObjectCannedACLPublicRead             ObjectCannedACL = "public-read"
	ObjectCannedACLPublicReadWrite        ObjectCannedACL = "public-read-write"
	ObjectCannedACLAuthenticatedRead      ObjectCannedACL = "authenticated-read"
	ObjectCannedACLAwsExecRead            ObjectCannedACL = "aws-exec-read"
	ObjectCannedACLBucketOwnerRead        ObjectCannedACL = "bucket-owner-read"
	ObjectCannedACLBucketOwnerFullControl ObjectCannedACL = "bucket-owner-full-control"
)

// Values returns all known values for ObjectCannedACL. Note that this can be
// expanded in the future, and so it is only as up to date as the client.
//
// The ordering of this slice is not guaranteed to be stable across updates.
func (ObjectCannedACL) Values() []ObjectCannedACL {
	return []ObjectCannedACL{
		"private",
		"public-read",
		"public-read-write",
		"authenticated-read",
		"aws-exec-read",
		"bucket-owner-read",
		"bucket-owner-full-control",
	}
}

// ObjectLockLegalHoldStatus specifies whether a legal hold will be applied to this object. For more
// information about S3 Object Lock, see [Object Lock] in the Amazon S3 User Guide.
type ObjectLockLegalHoldStatus string

// Enum values for ObjectLockLegalHoldStatus
const (
	ObjectLockLegalHoldStatusOn  ObjectLockLegalHoldStatus = "ON"
	ObjectLockLegalHoldStatusOff ObjectLockLegalHoldStatus = "OFF"
)

// ObjectLockMode is the Object Lock mode that you want to apply to this object.
type ObjectLockMode string

// Enum values for ObjectLockMode
const (
	ObjectLockModeGovernance ObjectLockMode = "GOVERNANCE"
	ObjectLockModeCompliance ObjectLockMode = "COMPLIANCE"
)

// RequestPayer confirms that the requester knows that they will be charged for the request.
// Bucket owners need not specify this parameter in their requests. If either the
// source or destination S3 bucket has Requester Pays enabled, the requester will
// pay for corresponding charges to copy the object. For information about
// downloading objects from Requester Pays buckets, see [Downloading Objects in Requester Pays Buckets]in the Amazon S3 User
// Guide.
type RequestPayer string

// Enum values for RequestPayer
const (
	RequestPayerRequester RequestPayer = "requester"
)

// ServerSideEncryption indicates the server-side encryption algorithm that was used when you store this object
// in Amazon S3 (for example, AES256 , aws:kms , aws:kms:dsse )
type ServerSideEncryption string

// Enum values for ServerSideEncryption
const (
	ServerSideEncryptionAes256     ServerSideEncryption = "AES256"
	ServerSideEncryptionAwsKms     ServerSideEncryption = "aws:kms"
	ServerSideEncryptionAwsKmsDsse ServerSideEncryption = "aws:kms:dsse"
)

// StorageClass specifies class to store newly created
// objects, which has default value of STANDARD. For more information, see
// [Storage Classes] in the Amazon S3 User Guide.
type StorageClass string

// Enum values for StorageClass
const (
	StorageClassStandard           StorageClass = "STANDARD"
	StorageClassReducedRedundancy  StorageClass = "REDUCED_REDUNDANCY"
	StorageClassStandardIa         StorageClass = "STANDARD_IA"
	StorageClassOnezoneIa          StorageClass = "ONEZONE_IA"
	StorageClassIntelligentTiering StorageClass = "INTELLIGENT_TIERING"
	StorageClassGlacier            StorageClass = "GLACIER"
	StorageClassDeepArchive        StorageClass = "DEEP_ARCHIVE"
	StorageClassOutposts           StorageClass = "OUTPOSTS"
	StorageClassGlacierIr          StorageClass = "GLACIER_IR"
	StorageClassSnow               StorageClass = "SNOW"
	StorageClassExpressOnezone     StorageClass = "EXPRESS_ONEZONE"
)

// CompletedPart includes details of the parts that were uploaded.
type CompletedPart struct {

	// The base64-encoded, 32-bit CRC32 checksum of the object. This will only be
	// present if it was uploaded with the object. When you use an API operation on an
	// object that was uploaded using multipart uploads, this value may not be a direct
	// checksum value of the full object. Instead, it's a calculation based on the
	// checksum values of each individual part. For more information about how
	// checksums are calculated with multipart uploads, see [Checking object integrity]in the Amazon S3 User
	// Guide.
	//
	// [Checking object integrity]: https://docs.aws.amazon.com/AmazonS3/latest/userguide/checking-object-integrity.html#large-object-checksums
	ChecksumCRC32 *string

	// The base64-encoded, 32-bit CRC32C checksum of the object. This will only be
	// present if it was uploaded with the object. When you use an API operation on an
	// object that was uploaded using multipart uploads, this value may not be a direct
	// checksum value of the full object. Instead, it's a calculation based on the
	// checksum values of each individual part. For more information about how
	// checksums are calculated with multipart uploads, see [Checking object integrity]in the Amazon S3 User
	// Guide.
	//
	// [Checking object integrity]: https://docs.aws.amazon.com/AmazonS3/latest/userguide/checking-object-integrity.html#large-object-checksums
	ChecksumCRC32C *string

	// The base64-encoded, 160-bit SHA-1 digest of the object. This will only be
	// present if it was uploaded with the object. When you use the API operation on an
	// object that was uploaded using multipart uploads, this value may not be a direct
	// checksum value of the full object. Instead, it's a calculation based on the
	// checksum values of each individual part. For more information about how
	// checksums are calculated with multipart uploads, see [Checking object integrity]in the Amazon S3 User
	// Guide.
	//
	// [Checking object integrity]: https://docs.aws.amazon.com/AmazonS3/latest/userguide/checking-object-integrity.html#large-object-checksums
	ChecksumSHA1 *string

	// The base64-encoded, 256-bit SHA-256 digest of the object. This will only be
	// present if it was uploaded with the object. When you use an API operation on an
	// object that was uploaded using multipart uploads, this value may not be a direct
	// checksum value of the full object. Instead, it's a calculation based on the
	// checksum values of each individual part. For more information about how
	// checksums are calculated with multipart uploads, see [Checking object integrity]in the Amazon S3 User
	// Guide.
	//
	// [Checking object integrity]: https://docs.aws.amazon.com/AmazonS3/latest/userguide/checking-object-integrity.html#large-object-checksums
	ChecksumSHA256 *string

	// Entity tag returned when the part was uploaded.
	ETag *string

	// Part number that identifies the part. This is a positive integer between 1 and
	// 10,000.
	//
	//   - General purpose buckets - In CompleteMultipartUpload , when a additional
	//   checksum (including x-amz-checksum-crc32 , x-amz-checksum-crc32c ,
	//   x-amz-checksum-sha1 , or x-amz-checksum-sha256 ) is applied to each part, the
	//   PartNumber must start at 1 and the part numbers must be consecutive.
	//   Otherwise, Amazon S3 generates an HTTP 400 Bad Request status code and an
	//   InvalidPartOrder error code.
	//
	//   - Directory buckets - In CompleteMultipartUpload , the PartNumber must start
	//   at 1 and the part numbers must be consecutive.
	PartNumber *int32
}

func (cp CompletedPart) mapCompletedPart() types.CompletedPart {
	return types.CompletedPart{
		ChecksumCRC32:  cp.ChecksumCRC32,
		ChecksumCRC32C: cp.ChecksumCRC32C,
		ChecksumSHA1:   cp.ChecksumSHA1,
		ChecksumSHA256: cp.ChecksumSHA256,
		ETag:           cp.ETag,
		PartNumber:     cp.PartNumber,
	}
}

// RequestCharged, if present, indicates that the requester was successfully charged for the
// request.
type RequestCharged string

// Enum values for RequestCharged
const (
	RequestChargedRequester RequestCharged = "requester"
)

// Metadata provides storing and reading metadata values. Keys may be any
// comparable value type. Get and set will panic if key is not a comparable
// value type.
//
// Metadata uses lazy initialization, and Set method must be called as an
// addressable value, or pointer. Not doing so may cause key/value pair to not
// be set.
type Metadata struct {
	values map[interface{}]interface{}
}
