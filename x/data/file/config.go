package file

const DefaultS3MaxUploadSize = 32 << 20

// S3Config contains the provider-specific settings required by S3Storage.
type S3Config struct {
	Endpoint      string
	Region        string
	Bucket        string
	AccessKey     string
	SecretKey     string
	UseSSL        bool
	PathStyle     bool  // Path-style (true) or virtual-hosted-style (false)
	MaxUploadSize int64 // Maximum bytes buffered for one Put; defaults to DefaultS3MaxUploadSize
}
