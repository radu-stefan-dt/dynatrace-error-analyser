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

package analyse

import (
	"fmt"
	"path/filepath"

	"github.com/radu-stefan-dt/dynatrace-error-analyser/pkg/config"
	"github.com/radu-stefan-dt/dynatrace-error-analyser/pkg/environment"
	"github.com/radu-stefan-dt/dynatrace-error-analyser/pkg/rest"
	"github.com/radu-stefan-dt/dynatrace-error-analyser/pkg/util"
	"github.com/spf13/afero"
)

func Analyse(dryRun bool, outputDir string, fs afero.Fs, environmentsFile string, configFile string, specificEnvironment string) error {
	environments, envErrors := environment.LoadEnvironmentList(specificEnvironment, environmentsFile, fs)
	configs, configErrors := config.LoadConfigList(configFile, fs)

	outputDir = filepath.Clean(outputDir)
	_ = outputDir
	_ = configs

	var deploymentErrors = make(map[string][]error)

	for i, err := range envErrors {
		configIssue := fmt.Sprintf("environmentfile-issue-%d", i)
		deploymentErrors[configIssue] = append(deploymentErrors[configIssue], err)
	}
	for i, err := range configErrors {
		configIssue := fmt.Sprintf("configurationfile-issue-%d", i)
		deploymentErrors[configIssue] = append(deploymentErrors[configIssue], err)
	}

	if !dryRun {
		for _, configuration := range configs {
			errors := execute(configuration, environments, outputDir)

			if len(errors) > 0 {
				deploymentErrors[configuration.GetId()] = errors
			}
		}
	}

	util.Log.Info("Deployment summary:")
	for _, errors := range deploymentErrors {
		if dryRun {
			util.Log.Error("Validation of environment failed. Found %d error(s)\n", len(errors))
			util.PrintErrors(errors)
		} else {
			util.Log.Error("Analysis of environment failed with error!\n")
			util.PrintErrors(errors)
		}
	}

	if dryRun {
		if len(deploymentErrors) > 0 {
			return fmt.Errorf("errors during validation! check log")
		} else {
			util.Log.Info("Validation finished without errors")
			return nil
		}
	} else {
		if len(deploymentErrors) > 0 {
			return fmt.Errorf("errors during execution! check log")
		} else {
			util.Log.Info("Execution finished without errors")
			return nil
		}
	}
}

func execute(config config.Config, environments map[string]environment.Environment, outputDir string) (errorList []error) {
	util.Log.Info("Running configuration %s", config.GetId())

	for _, env := range config.GetEnvironments() {
		util.Log.Info("\tAnalysing environment %s", env)

		environment := environments[env]
		var client rest.DynatraceClient

		apiToken, err := environment.GetToken()
		if err != nil {
			return append(errorList, err)
		}

		client, err = rest.NewDynatraceClient(environment.GetEnvironmentUrl(), apiToken)
		if err != nil {
			return append(errorList, err)
		}

		environmentErrors, err := client.FetchErrors(config)
		if err != nil {
			return append(errorList, err)
		}
		for _, envErr := range environmentErrors {
			userSessions, err := client.FetchSessionsByError(config, envErr)
			if err != nil {
				return append(errorList, err)
			}

			util.Log.Debug(fmt.Sprintf("Loaded %d user sessions!", len(userSessions)))

			switch useCase := config.GetUseCase(); useCase {
			case "lost_orders":
				analyseLostOrders(client, config, environmentErrors)
			case "agent_hours":
				analyseAgentHours(client, config, environmentErrors)
			case "incurred_costs":
				analyseIncurredCosts(client, config, environmentErrors)
			default:
				return append(errorList, fmt.Errorf("unrecognised value for use case: %s", useCase))
			}
		}

	}

	return errorList
}

func analyseLostOrders(client rest.DynatraceClient, config config.Config, environmentErrors []string) error {
	return nil
}

func analyseAgentHours(client rest.DynatraceClient, config config.Config, environmentErrors []string) error {
	return nil
}

func analyseIncurredCosts(client rest.DynatraceClient, config config.Config, environmentErrors []string) error {
	return nil
}
