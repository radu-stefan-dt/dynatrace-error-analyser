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

	"github.com/radu-stefan-dt/dynatrace-error-analyser/pkg/environment"
	"github.com/radu-stefan-dt/dynatrace-error-analyser/pkg/util"
	"github.com/spf13/afero"
)

func Analyse(dryRun bool, outputDir string, fs afero.Fs, environmentsFile string, specificEnvironment string) error {
	environments, errors := environment.LoadEnvironmentList(specificEnvironment, environmentsFile, fs)

	outputDir = filepath.Clean(outputDir)
	_ = outputDir

	var deploymentErrors = make(map[string][]error)

	for i, err := range errors {
		configIssue := fmt.Sprintf("environmentfile-issue-%d", i)
		deploymentErrors[configIssue] = append(deploymentErrors[configIssue], err)
	}

	util.Log.Debug("Environments:")
	for _, env := range environments {
		util.Log.Debug(fmt.Sprintf("%#v", env))
	}

	for _, environment := range environments {
		// errors := execute(environment, projects, dryRun, workingDir, continueOnError)
		errors := []error{}
		if len(errors) > 0 {
			deploymentErrors[environment.GetId()] = errors
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
