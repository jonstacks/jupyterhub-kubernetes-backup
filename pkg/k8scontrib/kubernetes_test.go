package k8scontrib

import (
	"os"
	"testing"

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
