package core

import "github.com/spf13/afero"

// Filesystem is the project wide abstraction for dealing with the filesystem.
// We'll use the to easily swap our the real filesystem for an in-memory
// one during tests, so the user doesn't actually have to have permissions
// to write to the real paths on the host.
var Filesystem = &afero.Afero{Fs: afero.NewOsFs()}
