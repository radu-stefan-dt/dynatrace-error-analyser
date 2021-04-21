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

package config

import (
	"errors"
	"fmt"
)

type Config interface {
	GetId() string
	GetName() string
	GetUseCase() UseCase
	GetProperty(property string) interface{}
	GetEnvironments() []string
}

type UseCase string

const (
	LostOrders    UseCase = "lost_orders"
	AgentHours    UseCase = "agent_hours"
	IncurredCosts UseCase = "incurred_costs"
)

type configImpl struct {
	id           string
	name         string
	useCase      UseCase
	properties   map[string]interface{}
	environments []string
}

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

func newConfig(id string, properties map[string]interface{}) (Config, error) {
	var configName string
	var useCase UseCase
	var configEnvs []string
	configProps := make(map[string]interface{})

	for k, v := range properties {
		switch t := v.(type) {
		case string:
			switch k {
			case "name":
				configName = t
			case "use_case":
				var err error
				useCase, err = getValidUseCase(t)
				if err != nil {
					return nil, fmt.Errorf("%#v is not a valid configuration use case", t)
				}
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
			default:
				return nil, fmt.Errorf("invalid format for %s. only environments can be specified as a list", k)
			}
		default:
			return nil, fmt.Errorf("invalid format for configuration detail %s", k)
		}
	}

	// TODO: need a final check of all mandatory details for creating a valid config

	return NewConfiguration(id, configName, useCase, configProps, configEnvs), nil
}

func NewConfiguration(id string, configName string, useCase UseCase,
	configProps map[string]interface{}, configEnvs []string) Config {
	return &configImpl{
		id:           id,
		name:         configName,
		useCase:      useCase,
		properties:   configProps,
		environments: configEnvs,
	}
}

func getValidUseCase(uc string) (UseCase, error) {
	switch uc {
	case "lost_orders":
		return LostOrders, nil
	case "agent_hours":
		return AgentHours, nil
	case "incurred_costs":
		return IncurredCosts, nil
	default:
		return "", fmt.Errorf("%s is not a valid config type", uc)
	}
}

func isValidUseCase(uc UseCase) error {
	switch uc {
	case LostOrders, AgentHours, IncurredCosts:
		return nil
	}
	return errors.New("invalid config use case")
}

func (s *configImpl) GetId() string {
	return s.id
}

func (s *configImpl) GetEnvironments() []string {
	return s.environments
}

func (s *configImpl) GetName() string {
	return s.name
}

func (s *configImpl) GetProperty(property string) interface{} {
	prop, ok := s.properties[property]
	if !ok {
		return nil
	}
	return prop
}

func (s *configImpl) GetUseCase() UseCase {
	if err := isValidUseCase(s.useCase); err == nil {
		return s.useCase
	} else {
		return ""
	}
}
