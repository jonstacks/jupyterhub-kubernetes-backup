package backend

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/jonstacks/jupyterhub-kubernetes-backup/pkg/core"
	"github.com/peakgames/s3hash"
	"github.com/sirupsen/logrus"
)

type s3Uploader interface {
	Upload(*s3manager.UploadInput, ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error)
}

type s3Client interface {
	HeadObject(*s3.HeadObjectInput) (*s3.HeadObjectOutput, error)
}

// S3 is an AWS S3 backend for storing the files
type S3 struct {
	client   s3Client
	uploader s3Uploader
	bucket   string
	prefix   string
}

// NewS3 creates a new S3 backend that implement Backend.
func NewS3(sess *session.Session, bucket string, prefix string) S3 {
	backend := S3{
		client:   s3.New(sess),
		uploader: s3manager.NewUploader(sess),
		bucket:   bucket,
		prefix:   prefix,
	}
	return backend
}

// Save saves the files at path to the backend
func (s S3) Save(basePath string) error {
	return core.Filesystem.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			logrus.Debugf("Entering directory: %s", path)
			return nil
		}

		rel, err := filepath.Rel(basePath, path)
		if err != nil {
			return err
		}

		key := fmt.Sprintf("%s/%s", s.prefix, rel)

		if !s.isObjectDirty(key, path) {
			logrus.Infof("File '%s' is already up to date in S3 based on ETag", key)
			return nil
		}

		f, err := core.Filesystem.Open(path)
		if err != nil {
			logrus.Errorf("Error opening file '%s': %s", path, err.Error())
			return err
		}
		defer f.Close()

		s3Params := &s3manager.UploadInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(key),
			Body:   f,
		}

		logrus.Infof("[Local] '%s' -> [s3://%s] '%s'", path, s.bucket, key)
		_, err = s.uploader.Upload(s3Params)
		if err != nil {
			logrus.Errorf("Error uploading local file %s: %s", path, err.Error())
		}

		return err
	})
}

func (s S3) isObjectDirty(key string, path string) bool {
	input := &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}
	fileContents, err := core.Filesystem.ReadFile(path)
	if err != nil {
		return true
	}

	resp, err := s.client.HeadObject(input)
	if err != nil || resp == nil {
		// In case of an error, return true we'll re-upload the file
		return true
	}

	s3Etag := aws.StringValue(resp.ETag)
	localEtag, err := s3hash.Calculate(bytes.NewReader(fileContents), s3manager.DefaultUploadPartSize)
	if err != nil {
		return true
	}

	logrus.WithFields(logrus.Fields{
		"local.etag":    localEtag,
		"local.path":    path,
		"remote.etag":   s3Etag,
		"remote.key":    key,
		"remote.bucket": s.bucket,
	}).Debugf("Comparing ETags to find if object is dirty")
	return s3Etag != localEtag
}
