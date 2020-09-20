package core

import "github.com/sirupsen/logrus"

// FatalIfError will log a message at the fatal error and exit if the
// error is not nil
func FatalIfError(err error) {
	if err != nil {
		logrus.Fatal(err)
	}
}
