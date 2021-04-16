/**
 * @license
 * Copyright (C) 2021  Radu Stefan
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see: https://www.gnu.org/licenses/
 **/

package util

import (
	"os"
	"path/filepath"
	"time"

	"github.com/jcelliott/lumber"
)

// Log is the shared Lumber Logger logging to console and after calling SetupLogging also to file
var Log lumber.Logger = lumber.NewConsoleLogger(lumber.INFO)
var requestLogFile *os.File

// SetupLogging is used to initialize the shared file Logger once the necesary setup config is available
func SetupLogging(verbose bool) error {
	multiLog := lumber.NewMultiLogger()
	consoleLog := lumber.NewConsoleLogger(lumber.INFO)
	if verbose {
		consoleLog.Level(lumber.DEBUG)
	}
	multiLog.AddLoggers(consoleLog)

	if _, err := os.Stat(".logs"); os.IsNotExist(err) {
		err = os.Mkdir(".logs", 0777)

		if err != nil {
			FailOnError(err, "could not create directory")
		}
	}

	logName := ".logs" + string(os.PathSeparator) + time.Now().Format("20060102-150405") + ".log"
	fileLog, err := lumber.NewAppendLogger(logName)

	if err != nil {
		return err
	}

	fileLog.Level(lumber.DEBUG)
	multiLog.AddLoggers(fileLog)
	Log = multiLog

	return setupRequestLog()
}

func setupRequestLog() error {
	if logFilePath, found := os.LookupEnv("DERRAN_REQUEST_LOG"); found {
		logFilePath, err := filepath.Abs(logFilePath)

		if err != nil {
			return err
		}

		Log.Debug("request log activated at %s", logFilePath)
		handle, err := prepareLogFile(logFilePath)

		if err != nil {
			return err
		}

		requestLogFile = handle
	} else {
		Log.Debug("request log not activated")
	}

	return nil
}

func prepareLogFile(file string) (*os.File, error) {
	return os.OpenFile(file, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
}

func IsRequestLoggingActive() bool {
	return requestLogFile != nil
}
