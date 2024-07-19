package transfermanager

type ChecksumAlgorithm string

const (
	ChecksumCRC32 ChecksumAlgorithm = "CRC32"
	ChecksumCRC32C = "CRC32C"
	ChecksumSHA1 = "SHA1"
	ChecksumSHA256 = "SHA256"
)
