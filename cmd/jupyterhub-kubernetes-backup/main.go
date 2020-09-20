package main

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	backendprovider "github.com/jonstacks/jupyterhub-kubernetes-backup/pkg/backend"
	"github.com/jonstacks/jupyterhub-kubernetes-backup/pkg/config"
	"github.com/jonstacks/jupyterhub-kubernetes-backup/pkg/core"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

func init() {
	logrus.SetFormatter(config.GetLogFormatter())
	logrus.SetLevel(config.GetLogLevel())
}

func main() {
	variables := config.NewMissingVariables()
	variables.Check(config.LocalPath)

	backend := config.Get(config.Backend)
	if backend == "" {
		backend = "mock"
	}

	logrus.Infof("Using %s backend", backend)

	var bkend backendprovider.Backend

	switch strings.ToLower(backend) {
	case "s3":
		variables.Check(config.BackendS3Bucket, config.BackendS3Prefix, config.BackupUsername)
		core.FatalIfError(variables.Missing())

		sess := session.Must(session.NewSessionWithOptions(session.Options{
			SharedConfigState: session.SharedConfigEnable,
		}))

		prefix := config.GetS3UserPrefix()
		bkend = backendprovider.NewS3(
			sess,
			config.Get(config.BackendS3Bucket),
			prefix,
		)

	default:
		core.FatalIfError(variables.Missing())
		bkend = backendprovider.NewMock(afero.NewOsFs())
	}

	core.FatalIfError(bkend.Save(config.Get(config.LocalPath)))
	logrus.Infof("Successfully backed up %s", config.Get(config.LocalPath))
}
