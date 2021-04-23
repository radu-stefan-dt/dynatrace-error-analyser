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
	"sort"
	"time"

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
			util.Log.Info("\t\tAnalysisng error %s", envErr)
			userSessions, err := client.FetchSessionsByError(config, envErr)

			if err != nil {
				return append(errorList, err)
			}

			util.Log.Debug(fmt.Sprintf("\t\tLoaded %d user sessions!", len(userSessions)))
			results, err := analyseSessions(userSessions, envErr, config)

			if err != nil {
				return append(errorList, err)
			}

			util.Log.Debug("\t\tResults are in:\n %#v", results)
		}

	}

	return errorList
}

func analyseSessions(userSessions []interface{}, envErr string,
	config config.Config) (results map[string]interface{}, err error) {

	conversion := config.GetProperty("conversion").(string)
	useCase := config.GetUseCase()
	errorAndAbandon, errorAndConvert, convert := splitUserSessions(envErr, conversion, useCase, userSessions)

	totalWithError := len(errorAndAbandon) + len(errorAndConvert)
	util.Log.Info("\t\t\t%d users got the error", totalWithError)
	util.Log.Info("\t\t\t%d users got the error and abandoned", len(errorAndAbandon))

	stats := calculateAbandonStats(useCase, errorAndAbandon, convert)
	lostTimes := stats["lost_times"].([]int64)
	lostBaskets := stats["lost_baskets"].(float64)
	lostUsers := stats["lost_users"].(int)

	sort.Slice(lostTimes, func(a, b int) bool {
		return lostTimes[a] > lostTimes[b]
	})

	var (
		dateCheck     string
		dayCount      int = 1
		dateBreakdown []interface{}
	)

	for i, lostTime := range lostTimes {
		dateString := time.Unix(0, lostTime*int64(time.Millisecond)).Format("02 Jan")

		if i == 0 {
			dateBreakdown = append(dateBreakdown, []string{"Date", "Amount of users"})
			dateBreakdown = append(dateBreakdown, []interface{}{dateString, 1})
			dateCheck = dateString
		} else if dateString == dateCheck {
			dateBreakdown[dayCount].([]interface{})[1] = dateBreakdown[dayCount].([]interface{})[1].(int) + 1
		} else if dateString != dateCheck {
			dateBreakdown = append(dateBreakdown, []interface{}{dateString, 1})
			dateCheck = dateString
			dayCount++
		}
	}
	results = make(map[string]interface{})
	results["use_case"] = useCase
	results["impacted_users"] = totalWithError
	results["lost_users"] = stats["lost_users"]
	results["unconverted_users"] = len(errorAndAbandon)
	results["user_breakdown"] = [4]interface{}{
		[2]interface{}{"User type", "User volume"},
		[2]interface{}{"Mobile", stats["lost_mobile"]},
		[2]interface{}{"Desktop", stats["lost_desktop"]},
		[2]interface{}{"Tablet", stats["lost_tablet"]},
	}
	results["date_breakdown"] = dateBreakdown

	if useCase == "lost_basket" {
		var multiFactor int = 1
		var margin float64 = 15
		if config.GetProperty("multiplication_factor") != nil {
			multiFactor = config.GetProperty("multiplication_factor").(int)
		}
		if config.GetProperty("margin") != nil {
			margin = config.GetProperty("margin").(float64)
		}

		lostBaskets *= float64(multiFactor)
		results["lost_basket"] = int(lostBaskets)

		lostMoney := int(lostBaskets / (100.0 / margin))
		results["lost_money"] = lostMoney
		results["lost_money_14d"] = lostMoney * 2
		results["lost_money_21d"] = lostMoney * 3
		results["lost_money_28d"] = lostMoney * 4
	} else if useCase == "agent_hours" {
		usersCalling := config.GetProperty("users_calling_in").(int)
		callLength := config.GetProperty("length_of_call").(int)
		callCost := config.GetProperty("cost_of_call").(float64)

		lostAgentHoursMin := (lostUsers / 100) * usersCalling * callLength
		lostAgentHoursHr := float64(lostAgentHoursMin / 60)

		if callCost == 0 {
			hoursLostCostMoney := float64((lostUsers/100)*usersCalling) * callCost
			results["hours_lost_cost"] = hoursLostCostMoney
			results["hours_lost_cost_14d"] = hoursLostCostMoney * 2
			results["hours_lost_cost_21d"] = hoursLostCostMoney * 3
			results["hours_lost_cost_28d"] = hoursLostCostMoney * 4
		}

		results["lost_agent_hours"] = lostAgentHoursHr
		results["lost_agent_hours_14d"] = lostAgentHoursHr * 2
		results["lost_agent_hours_21d"] = lostAgentHoursHr * 3
		results["lost_agent_hours_28d"] = lostAgentHoursHr * 4
	} else if useCase == "incurred_costs" {
		errorCost := config.GetProperty("cost_of_error").(float64)
		costsIncurred := lostUsers * int(errorCost)

		results["costs_incurred"] = costsIncurred
		results["costs_incurred_14d"] = costsIncurred * 2
		results["costs_incurred_21d"] = costsIncurred * 3
		results["costs_incurred_28d"] = costsIncurred * 4
	}

	return results, nil
}

func splitUserSessions(envErr string, conversion string, useCase config.UseCase, userSessions []interface{}) (
	errorAndAbandon []interface{}, errorAndConvert []interface{}, convert []interface{}) {

	for i := 0; i < len(userSessions); i++ {
		var converted bool
		session := userSessions[i].([]interface{})
		sDetails := util.UnpackSession(string(useCase), session)
		sessionErr := sDetails["error"].(string)
		sessionActions := sDetails["actions"].([]string)

		for _, action := range sessionActions {
			if action == conversion {
				converted = true
			}
		}

		if sessionErr == envErr {
			if converted {
				errorAndConvert = append(errorAndConvert, session)
			} else {
				errorAndAbandon = append(errorAndAbandon, session)
			}
		} else {
			convert = append(convert, session)
		}
	}

	return errorAndAbandon, errorAndConvert, convert
}

func calculateAbandonStats(useCase config.UseCase, errorAndAbandon []interface{},
	convert []interface{}) (stats map[string]interface{}) {
	var (
		savedBaskets float64
		lostBaskets  float64
		savedUsers   int
		lostUsers    int
		lostMobile   int
		lostDesktop  int
		lostTablet   int
		lostTimes    []int64
	)
	for i := 0; i < len(errorAndAbandon); i++ {
		var saved bool

		session := errorAndAbandon[i].([]interface{})
		sDetails := util.UnpackSession(string(useCase), session)
		userId := sDetails["userId"].(string)
		startTime := sDetails["startTime"].(int64)
		browserType := sDetails["browserType"].(string)
		basketValue := sDetails["basketValue"].(float64)

		for i2 := 0; i2 < len(convert); i2++ {
			convertedSession := convert[i].([]interface{})
			csDetails := util.UnpackSession(string(useCase), convertedSession)
			convertedId := csDetails["userId"].(string)
			convertedStartTime := csDetails["startTime"].(int64)

			if convertedId == userId && convertedStartTime > startTime {
				saved = true
			}
		}

		if saved {
			savedUsers++

			if useCase == "lost_basket" {
				savedBaskets += basketValue
			}
		} else {
			lostUsers++
			lostTimes = append(lostTimes, startTime)

			if useCase == "lost_basket" {
				lostBaskets += basketValue
			}
			if browserType == "Mobile Browser" {
				lostMobile++
			} else if browserType == "Desktop Browser" {
				lostDesktop++
			} else if browserType == "Tablet Browser" {
				lostTablet++
			}
		}
	}

	stats = make(map[string]interface{})
	stats["lost_baskets"] = lostBaskets
	stats["saved_baskets"] = savedBaskets
	stats["saved_users"] = savedUsers
	stats["lost_users"] = lostUsers
	stats["lost_mobile"] = lostMobile
	stats["lost_desktop"] = lostDesktop
	stats["lost_tablet"] = lostTablet
	stats["lost_times"] = lostTimes

	return stats
}
