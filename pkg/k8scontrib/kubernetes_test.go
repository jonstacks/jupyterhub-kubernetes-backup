package k8scontrib

import (
	"os"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func withEnv(key, value string, f func()) {
	prevValue, exists := os.LookupEnv(key)
	os.Setenv(key, value)
	f()
	if exists {
		os.Setenv(key, prevValue)
	} else {
		os.Unsetenv(key)
	}
}

func TestNamespaceFromEnvironment(t *testing.T) {
	withEnv("POD_NAMESPACE", "my-test-namespace", func() {
		assert.Equal(t, "my-test-namespace", Namespace())
	})

	assert.Equal(t, "default", Namespace())
}

func TestNamespaceFromFile(t *testing.T) {
	fs = afero.Afero{Fs: afero.NewMemMapFs()}
	fs.WriteFile(NamespaceFile, []byte("my-namespace-from-file"), 0644)
	assert.Equal(t, "my-namespace-from-file", Namespace())
}
