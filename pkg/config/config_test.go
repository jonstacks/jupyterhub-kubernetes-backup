package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func withEnv(envMap map[string]string, f func()) {
	oldValues := make(map[string]string)
	for name, value := range envMap {
		// Save previous value for after function call
		oldValues[name] = os.Getenv(name)
		os.Setenv(name, value)
	}
	f()
	for name, value := range oldValues {
		os.Setenv(name, value)
	}
}
func TestGetS3UserPrefix(t *testing.T) {
	os.Setenv(BackendS3Prefix, "my/prefix")
	os.Setenv(BackupUsername, "user.name")
	assert.Equal(t, "my/prefix/user.name", GetS3UserPrefix())
}

func TestGetDefault(t *testing.T) {
	withEnv(map[string]string{
		Backend:          "my-backend",
		AwsDefaultRegion: "us-west-2",
	}, func() {
		assert.Equal(t, "us-west-2", GetDefault(AwsDefaultRegion, "us-east-1"))
		assert.Equal(t, "my-backend", GetDefault(Backend, "mock"))
		assert.Equal(t, "defaultValue", GetDefault(LocalPath, "defaultValue"))
	})
}
