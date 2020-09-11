package config

import (
	"fmt"
	"os"
	"strings"
)

// MissingVariables is used to store the name of missing environment variables.
type MissingVariables struct {
	missing map[string]bool
}

// NewMissingVariables creates and initializes a new missing variable list
func NewMissingVariables() *MissingVariables {
	return &MissingVariables{
		missing: make(map[string]bool),
	}
}

// Check checks the given variables to make sure they are supplied
func (mv *MissingVariables) Check(names ...string) {
	for _, name := range names {
		val := os.Getenv(name)
		mv.missing[name] = val == ""
	}
}

// Missing returns an error if any of the environment variables are missing.
// Otherwise, it returns nil
func (mv *MissingVariables) Missing() error {
	missing := make([]string, 0)
	for name, notPresent := range mv.missing {
		if notPresent {
			missing = append(missing, name)
		}
	}

	if len(missing) == 0 {
		return nil
	}
	return fmt.Errorf("Missing the following required environment variables: %s", strings.Join(missing, ", "))
}
