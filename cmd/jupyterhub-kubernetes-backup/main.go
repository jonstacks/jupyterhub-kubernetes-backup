package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	backendprovider "github.com/jonstacks/jupyterhub-kubernetes-backup/pkg/backend"
	"github.com/jonstacks/jupyterhub-kubernetes-backup/pkg/config"
	"github.com/spf13/afero"
)

// FatalIfError exits the program on error
func FatalIfError(err error) {
	if err != nil {
		log.Fatal(err.Error())
	}
}

func main() {
	variables := config.NewMissingVariables()
	variables.Check(config.LocalPath)

	backend := config.Get(config.Backend)
	if backend == "" {
		backend = "mock"
	}

	log.Printf("Using %s backend", backend)

	var bkend backendprovider.Backend

	switch strings.ToLower(backend) {
	case "s3":
		variables.Check(config.BackendS3Bucket, config.BackendS3Prefix, config.BackupUsername)
		FatalIfError(variables.Missing())

		sess := session.Must(session.NewSessionWithOptions(session.Options{
			SharedConfigState: session.SharedConfigEnable,
		}))

		prefix := fmt.Sprintf("%s/%s", config.Get(config.BackendS3Prefix), config.Get(config.BackupUsername))
		bkend = backendprovider.NewS3(
			sess,
			config.Get(config.BackendS3Bucket),
			prefix,
		)

	default:
		FatalIfError(variables.Missing())
		bkend = backendprovider.NewMock(afero.NewOsFs())
	}

	FatalIfError(bkend.Save(config.Get(config.LocalPath)))
	log.Printf("Successfully backed up %s", config.Get(config.LocalPath))
}
