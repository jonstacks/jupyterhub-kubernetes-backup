package backend

import (
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
)

// Mock is a mock backend that optionally logs the files it would save
type Mock struct {
	fs              *afero.Afero
	LogFileNames    bool
	LogFileContents bool
}

// NewMock creates and initializes a new Mock backend
func NewMock(fs afero.Fs) Mock {
	return Mock{
		fs:              &afero.Afero{Fs: fs},
		LogFileNames:    true,
		LogFileContents: true,
	}
}

// Save saves the content at the given path to the backend
func (m Mock) Save(basePath string) error {
	return m.fs.Walk(basePath, func(path string, info os.FileInfo, err error) error {
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

		log.Printf("Processing file %s", rel)
		return nil
	})
}
