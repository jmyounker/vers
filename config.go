package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"
)

type Config struct {
	Data           map[string]interface{} `json:"data"`
	Branches       []BranchConfig         `json:"branches"`
	DataFileFields []string               `json:"data-file"`
}

type BranchConfig struct {
	BranchPattern   string `json:"branch"`
	VersionTemplate string `json:"version"`
}

func (c *Config) HasData(name string) bool {
	_, ok := c.Data[name]
	return ok
}

func (c *Config) GetDataInt(name string) (int, error) {
	v, ok := c.Data[name]
	if !ok {
		return 0, fmt.Errorf("data field '%s' is not defined", name)
	}
	switch v.(type) {
	case int:
		return v.(int), nil
	case float64:
		return int(v.(float64)), nil
	case string:
		iv, err := strconv.Atoi(v.(string))
		if err != nil {
			return 0, fmt.Errorf("cannot convert '%s' to an int: %s", name, err.Error())
		}
		return iv, nil
	default:
		return 0, fmt.Errorf("'%s' is not an int", name)
	}
}

func (c *Config) GetDataString(name string) (string, error) {
	v, ok := c.Data[name]
	if !ok {
		return "", fmt.Errorf("data field '%s' is not defined", name)
	}
	switch v.(type) {
	case int:
		return strconv.Itoa(v.(int)), nil
	case float64:
		return strconv.Itoa(int(v.(float64))), nil
	case string:
		return v.(string), nil
	default:
		return "", fmt.Errorf("expected '%s' to be a string", name)
	}
}

func (c *Config) writeConfig(filename string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, data, 0664)
}

func readConfig(filename string) (*Config, error) {
	var config Config
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	if len(config.Branches) == 0 {
		return nil, errors.New("confing must contain at least one branch expressions")
	}
	for _, bc := range config.Branches {
		err := checkBranchConfig(bc)
		if err != nil {
			return nil, err
		}
	}
	return &config, nil
}

func checkBranchConfig(bc BranchConfig) error {
	if bc.BranchPattern == "" {
		return errors.New("branch pattern required")
	}
	if bc.VersionTemplate == "" {
		return errors.New("version template required")
	}
	_, err := regexp.Compile("^" + bc.BranchPattern + "$")
	if err != nil {
		return fmt.Errorf("branch pattern '%s' is malformed", bc.BranchPattern)
	}
	t, err := ParseString(bc.VersionTemplate)
	if err != nil {
		return fmt.Errorf("version template '%s' is malformed", bc.VersionTemplate)
	}
	err = ValidateTemplateAsVersion(t)
	if err != nil {
		return err
	}
	return nil
}

func (c *Config) getBranchConfig(branch string) (*BranchConfig, *map[string]string, error) {
	for _, bc := range c.Branches {
		ptrn := regexp.MustCompile("^" + bc.BranchPattern + "$")
		matches := ptrn.FindStringSubmatch(branch)
		if len(matches) == 0 {
			continue
		}
		params := map[string]string{}
		paramNames := ptrn.SubexpNames()[1:]
		paramValues := matches[1:]
		for i, name := range(paramNames) {
			params[name] = paramValues[i]
		}
		return &bc, &params, nil
	}
	return nil, nil, fmt.Errorf("no branch config matching branch '%s'", branch)
}

func ValidateTemplateAsVersion(t Template) error {
	for _, v := range t.Variables() {
		if v == "version" {
			return errors.New("{version} cannot be contained in the version template")
		}
	}
	return nil
}

