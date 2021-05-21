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
 * GNU General Public License for more detailc.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see: https://www.gnu.org/licenses/
 **/

// Package config defines a error analysis configuration
package config

import (
	"errors"
	"fmt"
)

type Config interface {
	GetId() string
	GetName() string
	GetUseCases() []UseCase
	GetProperty(property string) interface{}
	GetEnvironments() []string
	HasUseCase(string) bool
}

type UseCase string

const (
	LostBasket    UseCase = "lost_basket"
	AgentHours    UseCase = "agent_hours"
	IncurredCosts UseCase = "incurred_costs"
)

type configImpl struct {
	id           string
	name         string
	useCases     []UseCase
	properties   map[string]interface{}
	environments []string
}

// NewConfigurations creates config items from a map of details.
// Map is expected to be map[string]map[string]interface{} as would be unmarshalled from YAML
func NewConfigurations(maps map[string]map[string]interface{}) (map[string]Config, []error) {
	configs := make(map[string]Config)
	errors := make([]error, 0)

	for id, details := range maps {
		config, err := newConfig(id, details)
		_ = config
		if err != nil {
			errors = append(errors, err)
		} else {
			// create error instead of overwriting configs with the same ID
			if _, configAlreadyExists := configs[config.GetId()]; !configAlreadyExists {
				configs[config.GetId()] = config
			} else {
				errors = append(errors, fmt.Errorf("configuration `%s` is already defined, please use unique configuration names", config.GetId()))
			}
		}
	}

	return configs, errors
}

// newConfig does any validation and sanity checks required before creating a new config
func newConfig(id string, properties map[string]interface{}) (Config, error) {
	var configName string
	var useCases []UseCase
	var configEnvs []string
	configProps := make(map[string]interface{})

	for k, v := range properties {
		switch t := v.(type) {
		case string:
			switch k {
			case "name":
				configName = t
			default:
				return nil, fmt.Errorf("invalid property %s found", k)
			}
		case map[interface{}]interface{}:
			for k2, v2 := range t {
				switch t2 := k2.(type) {
				case string:
					configProps[t2] = v2
				default:
					return nil, fmt.Errorf("invalid format for property %#v. keys may only be strings", k2)
				}
			}
		case []interface{}:
			switch k {
			case "environments":
				for _, env := range t {
					configEnvs = append(configEnvs, fmt.Sprintf("%s", env))
				}
			case "use_cases":
				for _, uc := range t {
					useCase, err := getValidUseCase(uc.(string))
					if err != nil {
						return nil, err
					}
					useCases = append(useCases, useCase)
				}
			default:
				return nil, fmt.Errorf("invalid format for %q. only environments and use_cases can be specified as a list", k)
			}
		default:
			return nil, fmt.Errorf("invalid format for configuration detail %q", k)
		}
	}

	err := checkMandatoryProperties(useCases, configProps)
	if err != nil {
		return nil, err
	}

	return NewConfiguration(id, configName, useCases, configProps, configEnvs), nil
}

// getMandatoryProperties returns the mandatory properties required by each analysis use case
func getMandatoryProperties(uc UseCase) []string {
	props := []string{"error_prop", "conversion"}
	switch uc {
	case LostBasket:
		props = append(props, "basket_prop")
	case AgentHours:
		props = append(props, "users_calling_in", "length_of_call")
	case IncurredCosts:
		props = append(props, "cost_of_error")
	}

	return props
}

// NewConfiguration creates a new config from given details.
// Should be used with clean data only. Any checks are done in newConfig.
func NewConfiguration(id string, configName string, useCases []UseCase,
	configProps map[string]interface{}, configEnvs []string) Config {
	return &configImpl{
		id:           id,
		name:         configName,
		useCases:     useCases,
		properties:   configProps,
		environments: configEnvs,
	}
}

// checkMandatoryProperties checks a map of properties and their values against the
// mandatory requrieremtns of the provided use cases.
func checkMandatoryProperties(useCases []UseCase, props map[string]interface{}) error {
	for _, useCase := range useCases {
		mProps := getMandatoryProperties(useCase)
		for _, mp := range mProps {
			found := false
			for p := range props {
				if p == mp {
					found = true
				}
			}

			if !found {
				return fmt.Errorf("use case %s is missing mandatory properties", useCase)
			}
		}
	}

	return nil
}

// getValidUseCase converts a string into a UseCase
func getValidUseCase(uc string) (UseCase, error) {
	switch uc {
	case "lost_basket":
		return LostBasket, nil
	case "agent_hours":
		return AgentHours, nil
	case "incurred_costs":
		return IncurredCosts, nil
	default:
		return "", fmt.Errorf("%q is not a valid use case", uc)
	}
}

func isValidUseCase(uc UseCase) error {
	switch uc {
	case LostBasket, AgentHours, IncurredCosts:
		return nil
	}
	return errors.New("invalid config use case")
}

// HasUseCase checks if a configuration references the given use case
func (c *configImpl) HasUseCase(uc string) bool {
	for _, useCase := range c.useCases {
		if string(useCase) == uc {
			return true
		}
	}

	return false
}

// GetId returns a configuration's ID
func (c *configImpl) GetId() string {
	return c.id
}

// GetEnvironments returns the IDs of environments referenced by the config
func (c *configImpl) GetEnvironments() []string {
	return c.environments
}

// GetName returns a configuration's name
func (c *configImpl) GetName() string {
	return c.name
}

// GetProperty returns, if successful, the value of a configuration's given property
func (c *configImpl) GetProperty(property string) interface{} {
	prop, ok := c.properties[property]
	if !ok {
		return nil
	}

	switch property {
	case "error_prop", "application", "conversion", "basket_prop":
		switch p := prop.(type) {
		case string:
			return p
		default:
			return nil
		}
	case "multiplication_factor", "users_calling_in", "length_of_call":
		switch p := prop.(type) {
		case float64:
			return int(p)
		case int:
			return p
		default:
			return nil
		}
	case "margin", "cost_of_call", "cost_of_error":
		switch p := prop.(type) {
		case int:
			return float64(p)
		case float64:
			return p
		default:
			return nil
		}
	default:
		return nil
	}
}

// GetUseCases returns the use cases referenced by the config
func (c *configImpl) GetUseCases() []UseCase {
	for _, useCase := range c.useCases {
		if err := isValidUseCase(useCase); err != nil {
			return nil
		}
	}

	return c.useCases
}
