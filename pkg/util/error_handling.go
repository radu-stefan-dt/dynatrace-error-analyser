package util

import (
	"errors"
	"os"
)

func PrintErrors(errors []error) {
	for _, err := range errors {
		Log.Error("\t%s", err)
	}
}

func FailOnError(err error, msg string) {
	if err != nil {
		Log.Fatal(msg + ": " + err.Error())
		os.Exit(1)
	}
}

func CheckProperty(properties map[string]string, property string) (string, error) {

	prop, ok := properties[property]
	if !ok {
		return "", errors.New("Property " + property + " was not available")
	}
	return prop, nil
}

func CheckError(err error, msg string) bool {
	if err != nil {
		Log.Error(msg + ": " + err.Error())
		return true
	}
	return false
}
