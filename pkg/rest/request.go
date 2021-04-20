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
	"io"
	"io/ioutil"
	"net/http"
	"runtime"

	"github.com/google/uuid"
	"github.com/radu-stefan-dt/dynatrace-error-analyser/pkg/util"
	"github.com/radu-stefan-dt/dynatrace-error-analyser/pkg/version"
)

type Response struct {
	StatusCode int
	Body       []byte
	Headers    map[string][]string
}

func get(client *http.Client, url string, apiToken string) (Response, error) {
	req, err := request(http.MethodGet, url, apiToken)

	if err != nil {
		return Response{}, err
	}

	return executeRequest(client, req), nil
}

func request(method string, url string, apiToken string) (*http.Request, error) {
	return requestWithBody(method, url, nil, apiToken)
}

func requestWithBody(method string, url string, body io.Reader, apiToken string) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)

	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Api-Token "+apiToken)
	req.Header.Set("Content-type", "application/json")
	req.Header.Set("User-Agent", "Dynatrace Error Analyser/"+version.ErrorAnalyser+" "+(runtime.GOOS+" "+runtime.GOARCH))
	return req, nil
}

func executeRequest(client *http.Client, request *http.Request) Response {
	var requestId string
	if util.IsRequestLoggingActive() {
		requestId = uuid.NewString()
		err := util.LogRequest(requestId, request)

		if err != nil {
			util.Log.Warn("error while writing request log for id `%s`: %v", requestId, err)
		}
	}

	rateLimitStrategy := createRateLimitStrategy()

	response, err := rateLimitStrategy.executeRequest(util.NewTimelineProvider(), func() (Response, error) {
		resp, err := client.Do(request)
		if err != nil {
			util.Log.Error("HTTP Request failed with Error: " + err.Error())
			return Response{}, err
		}
		defer func() {
			err = resp.Body.Close()
		}()
		body, err := ioutil.ReadAll(resp.Body)

		if util.IsResponseLoggingActive() {
			err := util.LogResponse(requestId, resp)

			if err != nil {
				if requestId != "" {
					util.Log.Warn("error while writing response log for id `%s`: %v", requestId, err)
				} else {
					util.Log.Warn("error while writing response log: %v", requestId, err)
				}
			}
		}

		return Response{
			StatusCode: resp.StatusCode,
			Body:       body,
			Headers:    resp.Header,
		}, err
	})

	if err != nil {
		// TODO properly handle error
		return Response{}
	}
	return response
}
