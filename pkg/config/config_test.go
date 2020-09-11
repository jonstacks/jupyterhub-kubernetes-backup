package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetS3UserPrefix(t *testing.T) {
	os.Setenv(BackendS3Prefix, "my/prefix")
	os.Setenv(BackupUsername, "user.name")
	assert.Equal(t, "my/prefix/user.name", GetS3UserPrefix())
}
