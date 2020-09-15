package backend

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type s3Uploader interface {
	Upload(*s3manager.UploadInput, ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error)
}

// S3 is an AWS S3 backend for storing the files
type S3 struct {
	uploader s3Uploader
	bucket   string
	prefix   string
}

// NewS3 creates a new S3 backend that implement Backend.
func NewS3(sess *session.Session, bucket string, prefix string) S3 {
	backend := S3{
		uploader: s3manager.NewUploader(sess),
		bucket:   bucket,
		prefix:   prefix,
	}
	return backend
}

// Save saves the files at path to the backend
func (s3 S3) Save(basePath string) error {
	return filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			log.Printf("Entering directory: %s", path)
			return nil
		}

		rel, err := filepath.Rel(basePath, path)
		if err != nil {
			return err
		}

		f, err := os.Open(path)
		if err != nil {
			log.Printf("Error opening file '%s': %s", path, err.Error())
			return err
		}

		key := fmt.Sprintf("%s/%s", s3.prefix, rel)
		s3Params := &s3manager.UploadInput{
			Bucket: aws.String(s3.bucket),
			Key:    aws.String(key),
			Body:   f,
		}

		log.Printf("[Local] '%s' -> [s3://%s] '%s'", path, s3.bucket, key)
		_, err = s3.uploader.Upload(s3Params)
		if err != nil {
			log.Printf("Error uploading local file %s: %s", path, err.Error())
		}

		return err
	})
}
