package backend

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestMockImplementsBackend(t *testing.T) {
	assert.Implements(t, (*Backend)(nil), new(Mock))
}

func TestMockSavesWithoutReturningError(t *testing.T) {
	fs := afero.NewMemMapFs()
	appFs := afero.Afero{Fs: fs}

	assert.NoError(t, appFs.WriteFile("/dummypath/hello.txt", []byte("world"), 0644))
	assert.NoError(t, appFs.WriteFile("/dummypath/super/nested/file.ipdb", []byte("package a"), 0644))

	mock := NewMock(fs)
	assert.NoError(t, mock.Save("/dummypath"))
}
