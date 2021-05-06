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

func (s *environmentImpl) GetId() string {
	return s.id
}

func (s *environmentImpl) GetEnvironmentUrl() string {
	return s.environmentUrl
}

func (s *environmentImpl) GetName() string {
	return s.name
}

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
