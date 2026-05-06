package file

// S3Config contains the provider-specific settings required by S3Storage.
type S3Config struct {
	Endpoint          string
	Region            string
	Bucket            string
	AccessKey         string
	SecretKey         string
	UseSSL            bool
	PathStyle         bool // Path-style (true) or virtual-hosted-style (false)
	TempDir           string
	MaxSinglePutBytes int64 // Maximum object size for the current single PUT implementation; 0 uses the default.
}
