package k8scontrib

import (
	"os"
	"strings"

	"github.com/jonstacks/jupyterhub-kubernetes-backup/pkg/core"
)

// NamespaceFile is the path of the file that kubernetes stores the namespace in
const NamespaceFile = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"

// Namespace returns the namespace of the current running pod
func Namespace() string {
	if ns, ok := os.LookupEnv("POD_NAMESPACE"); ok {
		return ns
	}

	if data, err := core.Filesystem.ReadFile(NamespaceFile); err == nil {
		if ns := strings.TrimSpace(string(data)); len(ns) > 0 {
			return ns
		}
	}
	return "default"
}
