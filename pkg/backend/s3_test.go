package backend

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/stretchr/testify/assert"
)

type testS3Uploader struct {
	injectError bool
}

func (uploader testS3Uploader) Upload(*s3manager.UploadInput, ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {
	if uploader.injectError {
		return nil, fmt.Errorf("an injected error occurred")
	}
	return nil, nil
}

func TestS3ImplementsBackend(t *testing.T) {
	assert.Implements(t, (*Backend)(nil), new(S3))
}

func TestSaveReturnsNilWhenNoErrors(t *testing.T) {
	uploader := testS3Uploader{injectError: false}

	s3 := &S3{
		uploader: uploader,
		bucket:   "test-bucket",
		prefix:   "my-prefix",
	}
	wd, err := os.Getwd()
	assert.NoError(t, err)
	assert.NoError(t, s3.Save(wd))
}

func TestSaveReturnsErrorWhenUploadErrorOccurs(t *testing.T) {
	uploader := testS3Uploader{injectError: true}

	s3 := &S3{
		uploader: uploader,
		bucket:   "test-bucket",
		prefix:   "my-prefix",
	}
	wd, err := os.Getwd()
	assert.NoError(t, err)
	assert.Error(t, s3.Save(wd))
}
