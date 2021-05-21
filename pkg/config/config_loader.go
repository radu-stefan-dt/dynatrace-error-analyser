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

package config

import (
	"errors"
	"fmt"

	"github.com/radu-stefan-dt/dynatrace-error-analyser/pkg/util"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
)

// LoadConfigList parses the configurations YAML file and returns the parsed configs, mapped by their ID
func LoadConfigList(configFile string, fs afero.Fs) (configs map[string]Config, errorList []error) {
	if configFile == "" {
		errorList = append(errorList, errors.New("no configurations file provided"))
		return configs, errorList
	}

	configsFromFile, errorList := readConfigs(configFile, fs)

	if len(configsFromFile) == 0 {
		errorList = append(errorList, fmt.Errorf("no configurations loaded from file %s", configFile))
		return configs, errorList
	} else {
		configs = configsFromFile
	}

	return configs, errorList
}

func readConfigs(file string, fs afero.Fs) (map[string]Config, []error) {
	data, err := afero.ReadFile(fs, file)
	util.FailOnError(err, "Error while reading config file")

	if errs := util.CheckUniqueYamlKey(data, "configuration id"); len(errs) > 0 {
		return nil, errs
	}

	configMaps := make(map[string]map[string]interface{})
	err = yaml.Unmarshal(data, &configMaps)
	util.FailOnError(err, "Error while converting YAML from config file")

	return NewConfigurations(configMaps)
}
