package config

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
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

func TestGetLogLevel(t *testing.T) {
	tests := map[string]logrus.Level{
		"":      logrus.InfoLevel, // Default
		"PANIC": logrus.PanicLevel,
		"DEBUG": logrus.DebugLevel,
		"TRACE": logrus.TraceLevel,
		"ERROR": logrus.ErrorLevel,
		"INFO":  logrus.InfoLevel,
		"BAD":   logrus.InfoLevel,
	}
	for env, lvl := range tests {
		withEnv(map[string]string{
			LogLevel: env,
		}, func() {
			assert.Equal(t, lvl, GetLogLevel())
		})
	}
}

func TestGetLogFormatter(t *testing.T) {
	tests := map[string]logrus.Formatter{
		"":     &logrus.TextFormatter{}, // Default
		"TEXT": &logrus.TextFormatter{},
		"JSON": &logrus.JSONFormatter{},
	}

	for env, fmt := range tests {
		withEnv(map[string]string{
			LogFormat: env,
		}, func() {
			assert.Equal(t, fmt, GetLogFormatter())
		})
	}
}

func TestGetBackupPodNodeAffinityRequired(t *testing.T) {
	tests := map[string]bool{
		"":          true,
		"preferred": false,
		"PREFERRED": false,
		"required":  true,
		"REQUIRED":  true,
	}
	for env, expected := range tests {
		withEnv(map[string]string{
			BackupPodNodeAffinity: env,
		}, func() {
			assert.Equal(t, expected, GetBackupPodNodeAffinityRequired())
		})
	}
}
