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
	"errors"
	"os"
)

func PrintErrorsFromMap(errors map[string][]error) {
	for id, errs := range errors {
		Log.Error("\t%s", id)
		for _, err := range errs {
			Log.Error("\t\t%s", err)
		}
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
