package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMissingVariables(t *testing.T) {
	mv := NewMissingVariables()

	withEnv(map[string]string{
		Backend:          "my-backend",
		AwsAccessKeyID:   "my-access-key",
		AwsDefaultRegion: "my-default-region",
	}, func() {
		mv.Check(Backend, AwsAccessKeyID, AwsDefaultRegion)
		assert.NoError(t, mv.Missing())

		mv.Check(AwsSecretAccessKey)
		assert.Error(t, mv.Missing())
	})
}
