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
	"time"

	"github.com/radu-stefan-dt/dynatrace-error-analyser/pkg/util"
)

// rateLimitStrategy ensures that the concrete implementation of the rate limiting strategy can be hidden
// behind this interface
type rateLimitStrategy interface {
	executeRequest(timelineProvider util.TimelineProvider, callback func() (Response, error)) (Response, error)
}

// createRateLimitStrategy creates a rateLimitStrategy. In the future this can be extended to instantiate
// different rate limiting strategies based on e.g. environment variables. The current implementation
// always returns the strategy simpleSleepRateLimitStrategy, which suspends the current goroutine until
// the time in the rate limiting header 'X-RateLimit-Reset' is up.
func createRateLimitStrategy() rateLimitStrategy {
	return &simpleSleepRateLimitStrategy{}
}

// simpleSleepRateLimitStrategy, is a rate limiting strategy which suspends the current goroutine until
// the time in the rate limiting header 'X-RateLimit-Reset' is up.
// It has a min sleep duration of 5 seconds and a max sleep duration of one minute and performs maximal 5
// polling iterations before giving up.
type simpleSleepRateLimitStrategy struct{}

func (s *simpleSleepRateLimitStrategy) executeRequest(timelineProvider util.TimelineProvider, callback func() (Response, error)) (Response, error) {

	response, err := callback()
	if err != nil {
		return Response{}, err
	}

	maxIterationCount := 5
	currentIteration := 0

	for response.StatusCode == http.StatusTooManyRequests && currentIteration < maxIterationCount {

		limit, humanReadableTimestamp, timeInMicroseconds, err := s.extractRateLimitHeaders(response)
		if err != nil {
			return response, err
		}

		util.Log.Info("Rate limit of %d requests/min reached: Applying rate limit strategy (simpleSleepRateLimitStrategy, iteration: %d)", limit, currentIteration+1)
		util.Log.Info("simpleSleepRateLimitStrategy: Attempting to sleep until %s", humanReadableTimestamp)

		// Attention: this uses client time:
		now := timelineProvider.Now()

		// Attention: this uses server time:
		resetTime := util.ConvertMicrosecondsToUnixTime(timeInMicroseconds)

		// Attention: this mixes client and server time:
		sleepDuration := resetTime.Sub(now)
		util.Log.Debug("simpleSleepRateLimitStrategy: Calculated sleep duration of %f seconds...", sleepDuration.Seconds())

		// That's why we need plausible min/max wait time defaults:
		sleepDuration = s.applyMinMaxDefaults(sleepDuration)

		util.Log.Debug("simpleSleepRateLimitStrategy: Sleeping for %f seconds...", sleepDuration.Seconds())
		timelineProvider.Sleep(sleepDuration)
		util.Log.Debug("simpleSleepRateLimitStrategy: Slept for %f seconds", sleepDuration.Seconds())

		// Checking again:
		currentIteration++

		response, err = callback()
		if err != nil {
			return Response{}, err
		}
	}

	return response, nil
}

func (s *simpleSleepRateLimitStrategy) extractRateLimitHeaders(response Response) (limit string, humanReadableResetTimestamp string, resetTimeInMicroseconds int64, err error) {

	limitAsArray := response.Headers["X-RateLimit-Limit"]
	resetAsArray := response.Headers["X-RateLimit-Reset"]

	if limitAsArray == nil || limitAsArray[0] == "" {
		return "", "", 0, errors.New("rate limit header 'X-RateLimit-Limit' not found")
	}
	if resetAsArray == nil || resetAsArray[0] == "" {
		return "", "", 0, errors.New("rate limit header 'X-RateLimit-Reset' not found")
	}

	limit = limitAsArray[0]
	humanReadableResetTimestamp, resetTimeInMicroseconds, err = util.StringTimestampToHumanReadableFormat(resetAsArray[0])
	if err != nil {
		return "", "", 0, err
	}

	return limit, humanReadableResetTimestamp, resetTimeInMicroseconds, nil
}

func (s *simpleSleepRateLimitStrategy) applyMinMaxDefaults(sleepDuration time.Duration) time.Duration {

	minWaitTimeInNanoseconds := 5 * time.Second
	maxWaitTimeInNanoseconds := 1 * time.Minute

	if sleepDuration.Nanoseconds() < minWaitTimeInNanoseconds.Nanoseconds() {
		sleepDuration = minWaitTimeInNanoseconds
		util.Log.Debug("simpleSleepRateLimitStrategy: Reset sleep duration to %f seconds...", sleepDuration.Seconds())
	}
	if sleepDuration.Nanoseconds() > maxWaitTimeInNanoseconds.Nanoseconds() {
		sleepDuration = maxWaitTimeInNanoseconds
		util.Log.Debug("simpleSleepRateLimitStrategy: Reset sleep duration to %f seconds...", sleepDuration.Seconds())
	}
	return sleepDuration
}
