package util

import "os"

func FailOnError(err error, msg string) {
	if err != nil {
		Log.Fatal(msg + ": " + err.Error())
		os.Exit(1)
	}
}
