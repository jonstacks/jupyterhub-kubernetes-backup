package config

import "os"

// Default config constants
const (
	LocalPath      = "LOCAL_PATH"
	BackupUsername = "BACKUP_USERNAME"

	Backend         = "BACKEND"
	DefaultBackend  = "mock"
	BackendS3Bucket = "BACKEND_S3_BUCKET"
	BackendS3Prefix = "BACKEND_S3_PREFIX"

	AwsAccessKeyID     = "AWS_ACCESS_KEY_ID"
	AwsSecretAccessKey = "AWS_SECRET_ACCESS_KEY"
)

// Get gets the environment variable by name
func Get(name string) string {
	return os.Getenv(name)
}

// GetDefault gets an environment variable if it exists, otehrwise the default.
func GetDefault(name string, defaultValue string) string {
	if val, ok := os.LookupEnv(name); ok {
		return val
	}
	return defaultValue
}
