package k8scontrib

import (
	"os"
	"strings"

	"github.com/spf13/afero"
)

// NamespaceFile is the path of the file that kubernetes stores the namespace in
const NamespaceFile = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"

var fs = afero.Afero{Fs: afero.NewOsFs()}

// Namespace returns the namespace of the current running pod
func Namespace() string {
	if ns, ok := os.LookupEnv("POD_NAMESPACE"); ok {
		return ns
	}

	if data, err := fs.ReadFile(NamespaceFile); err == nil {
		if ns := strings.TrimSpace(string(data)); len(ns) > 0 {
			return ns
		}
	}
	return "default"
}
