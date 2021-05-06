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
	"os"
	"sort"
	"strings"
)

// Replace whatever path separator was used with the correct one for the current O/S
func ReplacePathSeparators(path string) (newPath string) {
	newPath = strings.ReplaceAll(path, "\\", string(os.PathSeparator))
	newPath = strings.ReplaceAll(newPath, "/", string(os.PathSeparator))

	return newPath
}

// Safe way to extract the required details from a user session, in the correct format, avoiding nil panics
func UnpackSession(isLostBasket bool, session []interface{}) (sessionDetails map[string]interface{}) {
	var (
		actions     []string
		sErr        string
		browserType string
		userId      string
		startTime   int64
		basketValue float64
	)

	if fmt.Sprintf("%T", session[0]) == "string" {
		userId = session[0].(string)
	}
	if fmt.Sprintf("%T", session[1]) == "string" {
		sErr = session[1].(string)
	}
	if fmt.Sprintf("%T", session[2]) == "float64" {
		startTime = int64(session[2].(float64))
	}
	rawActions := session[4].([]interface{})
	for i := 0; i < len(rawActions); i++ {
		var action string
		if fmt.Sprintf("%T", rawActions[i]) == "string" {
			action = rawActions[i].(string)
		}
		actions = append(actions, action)
	}

	if isLostBasket {
		if fmt.Sprintf("%T", session[5]) == "float64" {
			basketValue = session[5].(float64)
		}
		if fmt.Sprintf("%T", session[6]) == "string" {
			browserType = session[6].(string)
		}
	} else {
		if fmt.Sprintf("%T", session[5]) == "string" {
			browserType = session[5].(string)
		}
	}

	sessionDetails = make(map[string]interface{})
	sessionDetails["error"] = sErr
	sessionDetails["userId"] = userId
	sessionDetails["actions"] = actions
	sessionDetails["startTime"] = startTime
	sessionDetails["basketValue"] = basketValue
	sessionDetails["browserType"] = browserType

	return sessionDetails
}

// Sort a map by its values
type Pair struct {
	Key   string
	Value int
}

type PairList []Pair

func (p PairList) Len() int           { return len(p) }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p PairList) Less(i, j int) bool { return p[i].Value > p[j].Value }

func SortMapDesc(m map[string]int) PairList {
	p := make(PairList, len(m))
	i := 0
	for k, v := range m {
		p[i] = Pair{k, v}
		i++
	}
	sort.Sort(p)

	return p
}
