package backend

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/jonstacks/jupyterhub-kubernetes-backup/pkg/core"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"
)

type testS3Uploader struct {
	injectError bool
}

func (uploader *testS3Uploader) Upload(*s3manager.UploadInput, ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {
	if uploader.injectError {
		return nil, fmt.Errorf("an injected error occurred")
	}
	return nil, nil
}

type testS3Client struct {
	injectError bool
	responses   map[string]*s3.HeadObjectOutput
}

func (client *testS3Client) HeadObject(input *s3.HeadObjectInput) (*s3.HeadObjectOutput, error) {
	if client.responses != nil && input.Key != nil {
		output, ok := client.responses[aws.StringValue(input.Key)]
		if ok {
			return output, nil
		}
	}
	return nil, nil
}

type S3BackendSuite struct {
	suite.Suite

	bucket          string
	prefix          string
	localBackupPath string

	fullKeyPath   func(string) string
	fullLocalPath func(string) string

	noErrorS3Client   *testS3Client
	errorS3Client     *testS3Client
	noErrorS3Uploader *testS3Uploader
	errorS3Uploader   *testS3Uploader

	defaultClient     *S3
	uploadErrorClient *S3
}

func (suite *S3BackendSuite) SetupTest() {
	suite.bucket = "test-bucket"
	suite.prefix = "my-prefix"
	suite.localBackupPath = "/backup"

	suite.fullKeyPath = func(path string) string {
		return fmt.Sprintf("%s/%s", suite.prefix, path)
	}

	suite.fullLocalPath = func(path string) string {
		return fmt.Sprintf("%s/%s", suite.localBackupPath, path)
	}

	// Set up an in memory filesystem that the S3 Backend will use
	// This prevents collisions with any user files and having to clean
	// up the filesystem when we are done.
	core.Filesystem = &afero.Afero{Fs: afero.NewMemMapFs()}

	suite.NoError(core.Filesystem.MkdirAll("/backup/subfolder", 0755))

	files := map[string][]byte{
		suite.fullLocalPath("test1.txt"):           []byte("This is file 1\n"),
		suite.fullLocalPath("test2.txt"):           []byte("This is file 2\n"),
		suite.fullLocalPath("subfolder/test3.txt"): []byte("This is file 3\n"),
	}

	for filename, contents := range files {
		suite.NoError(core.Filesystem.WriteFile(filename, contents, 0755))
	}

	suite.noErrorS3Client = &testS3Client{
		injectError: false,
		responses: map[string]*s3.HeadObjectOutput{

			suite.fullKeyPath("test1.txt"):           {ETag: aws.String("\"88c16a56754e0f17a93d269ae74dde9b\"")},
			suite.fullKeyPath("test2.txt"):           {ETag: aws.String("")},
			suite.fullKeyPath("subfolder/test3.txt"): {ETag: aws.String("")},
		},
	}
	suite.errorS3Client = &testS3Client{injectError: true}
	suite.noErrorS3Uploader = &testS3Uploader{injectError: false}
	suite.errorS3Uploader = &testS3Uploader{injectError: true}

	suite.defaultClient = &S3{
		client:   suite.noErrorS3Client,
		uploader: suite.noErrorS3Uploader,
		bucket:   suite.bucket,
		prefix:   suite.prefix,
	}

	suite.uploadErrorClient = &S3{
		client:   suite.noErrorS3Client,
		uploader: suite.errorS3Uploader,
		bucket:   suite.bucket,
		prefix:   suite.prefix,
	}
}

func (suite *S3BackendSuite) TestImplementsBackend() {
	suite.Implements((*Backend)(nil), new(S3))
}

func (suite *S3BackendSuite) TestNewS3() {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	s3 := NewS3(sess, suite.bucket, suite.prefix)
	suite.Equal(s3.bucket, suite.bucket)
	suite.Equal(s3.prefix, suite.prefix)
}

func (suite *S3BackendSuite) TestSaveReturnsNilWhenNoErrors() {
	suite.NoError(suite.defaultClient.Save(suite.localBackupPath))
}

func (suite *S3BackendSuite) TestSaveReturnsErrorWhenUploadErrorOccurs() {
	suite.Error(suite.uploadErrorClient.Save(suite.localBackupPath))
}

func (suite *S3BackendSuite) TestSaveReturnsErrorWhenFileIsGivenInsteadOfDirectory() {
	suite.Error(suite.defaultClient.Save("/badpath"))
}

func (suite *S3BackendSuite) TestIsObjectDirtyReturnsFalseIfETagsMatch() {
	// test1.txt's ETag should match, so no new upload should be necessary.
	suite.False(suite.defaultClient.isObjectDirty(
		suite.fullKeyPath("test1.txt"),
		suite.fullLocalPath("test1.txt"),
	))
}

func (suite *S3BackendSuite) TestIsObjectDirty() {

	trueTestCases := [][2]string{
		{"notFound.txt", "test2.txt"},
		{"test2.txt", "notFound.txt"},
		{"test2.txt", "test2.txt"},
	}

	for _, tc := range trueTestCases {
		suite.True(
			suite.defaultClient.isObjectDirty(
				suite.fullKeyPath(tc[0]),
				suite.fullLocalPath(tc[1]),
			),
		)
	}
}

func TestS3BackendSuite(t *testing.T) {
	suite.Run(t, new(S3BackendSuite))
}
