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

package util

import (
	"fmt"
	"strconv"
	"time"
)

//go:generate mockgen -source=time.go -destination=time_mock.go -package=util TimelineProvider

// TimelineProvider abstracts away the time.Now() and time.Sleep(time.Duration) functions to make code unit-testable
// Whenever you need to get the current time, or want to pause the current goroutine (sleep), please consider using
// this interface
type TimelineProvider interface {

	// Now Returns the current (client-side) time in UTC
	Now() time.Time

	// NowMillis returns the current (client-side) time in UTC Milliseconds
	NowMillis() int64

	// Takes number of days and returns the UTC Milliseconds from that many days ago
	GetDaysBeforeMillis(days int) int64

	// Sleep suspends the current goroutine for the specified duration
	Sleep(duration time.Duration)
}

// NewTimelineProvider creates a new TimelineProvider
func NewTimelineProvider() TimelineProvider {
	return &defaultTimelineProvider{}
}

// defaultTimelineProvider is the default implementation of interface TimelineProvider
type defaultTimelineProvider struct{}

func (d *defaultTimelineProvider) Now() time.Time {
	nowInLocalTimeZone := time.Now()
	location, _ := time.LoadLocation("UTC")
	return nowInLocalTimeZone.In(location)
}

func (d *defaultTimelineProvider) NowMillis() int64 {
	nowInLocalTimeZone := time.Now()
	location, _ := time.LoadLocation("UTC")
	utcNanos := nowInLocalTimeZone.In(location).UnixNano()

	return utcNanos / 1_000_000
}

func (d *defaultTimelineProvider) GetDaysBeforeMillis(days int) int64 {
	nowInLocalTimeZone := time.Now()
	nowInLocalTimeZone = nowInLocalTimeZone.AddDate(0, 0, -days)
	location, _ := time.LoadLocation("UTC")
	utcNanos := nowInLocalTimeZone.In(location).UnixNano()

	return utcNanos / 1_000_000
}

func (d *defaultTimelineProvider) Sleep(duration time.Duration) {
	time.Sleep(duration)
}

// StringTimestampToHumanReadableFormat parses and sanity-checks a unix timestamp as string and returns it
// as int64 and a human-readable representation of it
func StringTimestampToHumanReadableFormat(unixTimestampAsString string) (humanReadable string, parsedTimestamp int64, err error) {

	result, err := strconv.ParseInt(unixTimestampAsString, 10, 64)
	if err != nil {
		return "", 0, fmt.Errorf(
			"%s is not a valid unix timestamp",
			unixTimestampAsString,
		)
	}

	unixTimeUTC := time.Unix(result, 0)
	return unixTimeUTC.Format(time.RFC3339), result, nil
}

// ConvertMicrosecondsToUnixTime converts the UTC time in microseconds to a time.Time struct (unix time)
func ConvertMicrosecondsToUnixTime(timeInMicroseconds int64) time.Time {

	resetTimeInSeconds := timeInMicroseconds / 1000000
	resetTimeRemainderInNanoseconds := (timeInMicroseconds % 1000000) * 1000

	return time.Unix(resetTimeInSeconds, resetTimeRemainderInNanoseconds)
}
