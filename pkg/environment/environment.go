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

// Package environment defines a Dynatrace Environment
package environment

import (
	"fmt"
	"os"
	"strings"

	"github.com/radu-stefan-dt/dynatrace-error-analyser/pkg/util"
)

type Environment interface {
	GetId() string
	GetEnvironmentUrl() string
	GetToken() (string, error)
	GetName() string
}

type environmentImpl struct {
	id             string
	name           string
	environmentUrl string
	envTokenName   string
	envToken       string
}

// NewEnvironments creates one or more environments from a map of details.
// The resulting Environments are mapped by the unique ID as the key. Different errors that
// may occur during parsing and validation are collated and returned as a list of errors.
func NewEnvironments(maps map[string]map[string]string) (map[string]Environment, []error) {

	environments := make(map[string]Environment)
	errors := make([]error, 0)

	for id, details := range maps {
		environment, err := newEnvironment(id, details)
		if err != nil {
			errors = append(errors, err)
		} else {
			// create error instead of overwriting environments with same IDs
			if _, environmentAlreadyExists := environments[environment.GetId()]; !environmentAlreadyExists {
				environments[environment.GetId()] = environment
			} else {
				errors = append(errors, fmt.Errorf("environment `%s` is already defined, please use unique environment names", environment.GetId()))
			}
		}
	}

	return environments, errors
}

func newEnvironment(id string, properties map[string]string) (Environment, error) {

	environmentName, nameErr := util.CheckProperty(properties, "name")
	environmentUrl, urlErr := util.CheckProperty(properties, "env-url")
	envTokenName, tokenNameErr := util.CheckProperty(properties, "env-token-name")
	envToken, tokenErr := util.CheckProperty(properties, "env-token")

	// TODO: Improve error handling
	if nameErr != nil || urlErr != nil || (tokenErr != nil && tokenNameErr != nil) {
		return nil, fmt.Errorf("failed to parse config for environment %s. issues found:\n \t%s\n \t%s\n \t%s\n \t%s\n)", id, nameErr, urlErr, tokenNameErr, tokenErr)
	}

	return NewEnvironment(id, environmentName, environmentUrl, envTokenName, envToken), nil
}

// NewEnvironment creates a new Environment based on mandatory details.
// It should only be used with clean data. Any pre validation and checking should be done in newEnvironment.
func NewEnvironment(id string, name string, environmentUrl string, envTokenName string, envToken string) Environment {
	environmentUrl = strings.TrimSuffix(environmentUrl, "/")

	return &environmentImpl{
		id:             id,
		name:           name,
		environmentUrl: environmentUrl,
		envTokenName:   envTokenName,
		envToken:       envToken,
	}
}

// GetId returns an environment's ID
func (s *environmentImpl) GetId() string {
	return s.id
}

// GetEnvironmentUrl returns an environment's URL
func (s *environmentImpl) GetEnvironmentUrl() string {
	return s.environmentUrl
}

func (s *environmentImpl) GetName() string {
	return s.name
}

// GetToken returns the value of the API Token associated with the Dynatrace environment.
func (s *environmentImpl) GetToken() (string, error) {
	if s.envToken != "" {
		return s.envToken, nil
	} else {
		value := os.Getenv(s.envTokenName)
		if value == "" {
			return value, fmt.Errorf("no token value found, and environment variable " + s.envTokenName + " also not found")
		}
		return value, nil
	}
}
