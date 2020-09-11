package config

import "os"

// Default config constants
const (
	LocalPath = "LOCAL_PATH"

	Backend         = "BACKEND"
	DefaultBackend  = "S3"
	BackendS3Bucket = "BACKEND_S3_BUCKET"
	BackendS3Prefix = "BACKEND_S3_PREFIX"
)

// Get gets the environment variable by name
func Get(name string) string {
	return os.Getenv(name)
}
