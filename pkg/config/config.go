package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

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
	AwsDefaultRegion   = "AWS_DEFAULT_REGION"

	LogLevel  = "LOG_LEVEL"
	LogFormat = "LOG_FORMAT"
)

var (
	defaultLogLevel     = logrus.InfoLevel
	defaultLogFormatter = &logrus.TextFormatter{}
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

// GetS3UserPrefix should only be called after validating the config. It returns
// the s3 prefix for uploads.
func GetS3UserPrefix() string {
	return fmt.Sprintf("%s/%s", Get(BackendS3Prefix), Get(BackupUsername))
}

// GetLogLevel returns the log level from the environment
func GetLogLevel() logrus.Level {
	lvl := os.Getenv(LogLevel)
	if lvl != "" {
		level, err := logrus.ParseLevel(lvl)
		if err == nil {
			return level
		}
		logrus.Errorf("Error parsing LOG_LEVEL='%s' from environment. Defaulting to INFO", lvl)
	}
	return defaultLogLevel
}

// GetLogFormatter returns the log formatter from the environment
func GetLogFormatter() logrus.Formatter {
	fmt := os.Getenv(LogFormat)
	switch strings.ToLower(fmt) {
	case "json":
		return &logrus.JSONFormatter{}
	}
	return defaultLogFormatter
}
