package environment

import (
	"errors"
	"fmt"

	"github.com/radu-stefan-dt/dynatrace-error-analyser/pkg/util"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
)

func LoadEnvironmentList(specificEnvironment string, environmentsFile string, fs afero.Fs) (environments map[string]Environment, errorList []error) {

	if environmentsFile == "" {
		errorList = append(errorList, errors.New("no environments file provided"))
		return environments, errorList
	}

	environmentsFromFile, errorList := readEnvironments(environmentsFile, fs)

	if len(environmentsFromFile) == 0 {
		errorList = append(errorList, fmt.Errorf("no environments loaded from file %s", environmentsFile))
		return environments, errorList
	}

	if specificEnvironment != "" {
		if environmentsFromFile[specificEnvironment] == nil {
			errorList = append(errorList, fmt.Errorf("environment %s not found in file %s", specificEnvironment, environmentsFile))
			return environments, errorList
		}

		environments = make(map[string]Environment)
		environments[specificEnvironment] = environmentsFromFile[specificEnvironment]
	} else {
		environments = environmentsFromFile
	}

	return environments, errorList
}

// readEnvironments reads the yaml file for the environments and returns the parsed environments
func readEnvironments(file string, fs afero.Fs) (map[string]Environment, []error) {

	data, err := afero.ReadFile(fs, file)
	util.FailOnError(err, "Error while reading file")

	environmentMaps := make(map[string]map[string]string)
	err = yaml.Unmarshal(data, &environmentMaps)
	util.FailOnError(err, "Error while converting file")

	return NewEnvironments(environmentMaps)
}
