module github.com/jonstacks/jupyterhub-kubernetes-backup

go 1.13

require (
	github.com/aws/aws-sdk-go v1.34.21
	github.com/peakgames/s3hash v0.1.1
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/afero v1.3.5
	github.com/stretchr/testify v1.5.1
	golang.org/x/net v0.0.0-20200822124328-c89045814202 // indirect
	k8s.io/api v0.19.0
	k8s.io/apimachinery v0.19.0
	k8s.io/client-go v0.19.0
)
