package backend

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestS3ImplementsBackend(t *testing.T) {
	assert.Implements(t, (*Backend)(nil), new(S3))
}
