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
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strings"

	"github.com/radu-stefan-dt/dynatrace-error-analyser/pkg/config"
	"github.com/radu-stefan-dt/dynatrace-error-analyser/pkg/util"
)

type DynatraceClient interface {
	// Retrieves a list of error names as captured by the string property referenced in config.yaml
	FetchErrors(config config.Config) (environmentErrors []string, err error)

	// Retrieves user session data for sessions that encountered given error
	FetchSessionsByError(config config.Config, envErr string) (sessions []interface{}, err error)
}

type dynatraceClientImpl struct {
	environmentUrl string
	token          string
	client         *http.Client
}

const (
	userSessionsTableAPI string = "/api/v1/userSessionQueryLanguage/table"
)

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

func (d *dynatraceClientImpl) FetchErrors(config config.Config) (environmentErrors []string, err error) {
	application := fmt.Sprintf("%s", config.GetProperty("application"))
	errorProp := fmt.Sprintf("%s", config.GetProperty("error_prop"))

	if application != "" {
		application = "useraction.application IS \"" + application + "\" AND "
	}

	timer := util.NewTimelineProvider()
	now := timer.NowMillis()
	dayAgo := timer.GetDaysBeforeMillis(1)

	query := "SELECT DISTINCT stringProperties." + errorProp + ", count(*) FROM usersession WHERE " + application + "stringProperties." + errorProp + " IS NOT NULL"
	params := url.Values{}
	params.Add("query", query)
	params.Add("startTimestamp", fmt.Sprintf("%d", dayAgo))
	params.Add("endTimestamp", fmt.Sprintf("%d", now))
	params.Add("addDeepLinkFields", "false")
	params.Add("explain", "false")
	fullUrl := d.environmentUrl + userSessionsTableAPI + "?" + params.Encode()
	fmt.Println(fullUrl)

	response, err := get(d.client, fullUrl, d.token)

	if err != nil {
		return nil, err
	}

	var m map[string]interface{}
	if err := json.Unmarshal(response.Body, &m); err != nil {
		return nil, err
	}

	values := m["values"].([]interface{})

	for i := range values {
		value := values[i].([]interface{})
		// TODO: Remove this
		// Added limit of 10 for testing purposes
		if i == 10 {
			break
		}
		errorName := value[0].(string)
		errorCount := value[1].(float64)

		if errorCount < 4000 {
			environmentErrors = append(environmentErrors, errorName)
		}
	}

	return environmentErrors, nil
}

func (d *dynatraceClientImpl) FetchSessionsByError(config config.Config,
	envErr string) (sessions []interface{}, err error) {

	application := config.GetProperty("application").(string)
	errorProp := config.GetProperty("error_prop").(string)
	conversion := config.GetProperty("conversion").(string)
	useCase := config.GetUseCase()

	if application != "" {
		application = "useraction.application IS \"" + application + "\" AND "
	}

	timer := util.NewTimelineProvider()
	now := timer.NowMillis()
	sevenDaysAgo := timer.GetDaysBeforeMillis(7)

	query := "SELECT count(*) FROM usersession WHERE " + application + "useraction.name IS \"" + conversion + "\" OR stringProperties." + errorProp + " IS \"" + envErr + "\" LIMIT 5000"
	params := url.Values{}
	params.Add("query", query)
	params.Add("startTimestamp", fmt.Sprintf("%d", sevenDaysAgo))
	params.Add("endTimestamp", fmt.Sprintf("%d", now))
	params.Add("addDeepLinkFields", "false")
	params.Add("explain", "false")
	fullUrl := d.environmentUrl + userSessionsTableAPI + "?" + params.Encode()

	response, err := get(d.client, fullUrl, d.token)

	if err != nil {
		return nil, err
	}

	var m map[string]interface{}
	if err := json.Unmarshal(response.Body, &m); err != nil {
		return nil, err
	}

	extrapolation := m["extrapolationLevel"].(float64)
	values := m["values"].([]interface{})
	count := int(values[0].([]interface{})[0].(float64))

	util.Log.Debug(fmt.Sprintf("Extrapolation level: %.0f and %d sessions returned.", extrapolation, count))

	if extrapolation > 1 || count > 4999 {
		var numberOfQueriesEx int
		var numberOfQueriesLen int
		var numberOfQueries int

		if extrapolation > 1 {
			numberOfQueriesEx = int(extrapolation) / 2
		}
		if count > 4999 {
			numberOfQueriesLen = int(math.Round(float64(count / 5000)))
		}
		numberOfQueries = int(math.Max(float64(numberOfQueriesEx), float64(numberOfQueriesLen)))
		util.Log.Info("Will split results in %d queries", numberOfQueries)

		interval := (now - sevenDaysAgo) / int64(numberOfQueries)

		for i := 0; i < numberOfQueries; i++ {
			startTime := sevenDaysAgo + int64(i)*interval
			endTime := sevenDaysAgo + int64(i+1)*interval
			var query string

			if useCase == "lost_basket" {
				basketProp := fmt.Sprintf("%s", config.GetProperty("basket_prop"))
				query += "SELECT internalUserId, stringProperties." + errorProp + ", startTime, endTime, useraction.name, doubleProperties." + basketProp
				query += ", browserType FROM usersession WHERE " + application + "useraction.name IS \"" + conversion + "\" OR stringProperties."
				query += errorProp + " IS \"" + envErr + "\" LIMIT 5000"
			} else {
				query += "SELECT internalUserId, stringProperties." + errorProp + ", startTime, endTime, useraction.name, browserType FROM usersession WHERE "
				query += application + "useraction.name IS \"" + conversion + "\" OR stringProperties." + errorProp + " IS \"" + envErr + "\" LIMIT 5000"
			}

			params := url.Values{}
			params.Add("query", query)
			params.Add("startTimestamp", fmt.Sprintf("%d", startTime))
			params.Add("endTimestamp", fmt.Sprintf("%d", endTime))
			params.Add("addDeepLinkFields", "false")
			params.Add("explain", "false")
			fullUrl = d.environmentUrl + userSessionsTableAPI + "?" + params.Encode()

			response, err := get(d.client, fullUrl, d.token)

			if err != nil {
				return nil, err
			}

			var m map[string]interface{}
			if err := json.Unmarshal(response.Body, &m); err != nil {
				return nil, err
			}
			values := m["values"].([]interface{})
			sessions = append(sessions, values...)
		}
	} else {
		var query string

		if useCase == "lost_basket" {
			basketProp := fmt.Sprintf("%s", config.GetProperty("basket_prop"))
			query += "SELECT internalUserId, stringProperties." + errorProp + ", startTime, endTime, useraction.name, doubleProperties." + basketProp
			query += ", browserType FROM usersession WHERE " + application + "useraction.name IS \"" + conversion + "\" OR stringProperties." + errorProp
			query += " IS \"" + envErr + "\" LIMIT 5000"
		} else {
			query += "SELECT internalUserId, stringProperties." + errorProp + ", startTime, endTime, useraction.name, browserType FROM usersession WHERE "
			query += application + "useraction.name IS \"" + conversion + "\" OR stringProperties." + errorProp + " IS \"" + envErr + "\" LIMIT 5000"
		}

		params := url.Values{}
		params.Add("query", query)
		params.Add("startTimestamp", fmt.Sprintf("%d", sevenDaysAgo))
		params.Add("endTimestamp", fmt.Sprintf("%d", now))
		params.Add("addDeepLinkFields", "false")
		params.Add("explain", "false")
		fullUrl = d.environmentUrl + userSessionsTableAPI + "?" + params.Encode()

		response, err := get(d.client, fullUrl, d.token)

		if err != nil {
			return nil, err
		}

		var m map[string]interface{}
		if err := json.Unmarshal(response.Body, &m); err != nil {
			return nil, err
		}
		values = m["values"].([]interface{})
		sessions = append(sessions, values...)
	}

	return sessions, nil
}
