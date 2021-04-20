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

package rest

import (
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/radu-stefan-dt/dynatrace-error-analyser/pkg/util"
)

type DynatraceClient interface {
}

type dynatraceClientImpl struct {
	environmentUrl string
	token          string
	client         *http.Client
}

// NewDynatraceClient creates a new DynatraceClient
func NewDynatraceClient(environmentUrl, token string) (DynatraceClient, error) {

	if environmentUrl == "" {
		return nil, errors.New("no environment url")
	}

	if token == "" {
		return nil, errors.New("no token")
	}

	parsedUrl, err := url.ParseRequestURI(environmentUrl)
	if err != nil {
		return nil, errors.New("environment url " + environmentUrl + " was not valid")
	}

	if parsedUrl.Scheme != "https" {
		return nil, errors.New("environment url " + environmentUrl + " was not valid")
	}

	if !isNewDynatraceTokenFormat(token) {
		util.Log.Warn("You used an old token format. Please consider switching to the new 1.205+ token format.")
		util.Log.Warn("More information: https://www.dynatrace.com/support/help/dynatrace-api/basics/dynatrace-api-authentication/#-dynatrace-version-1205--token-format")
	}

	return &dynatraceClientImpl{
		environmentUrl: environmentUrl,
		token:          token,
		client:         &http.Client{},
	}, nil
}

func isNewDynatraceTokenFormat(token string) bool {
	return strings.HasPrefix(token, "dt0c01.") && strings.Count(token, ".") == 2
}
