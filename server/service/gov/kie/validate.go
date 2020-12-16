package kie

import (
	"fmt"
)

type Validator struct {
}

var methodSet map[string]bool

func (d *Validator) Validate(kind string, spec interface{}) error {
	switch kind {
	case "match-group":
		return matchValidate(spec)
	case "retry":
		return retryValidate(spec)
	case "rate-limiting":
		return rateLimitingValidate(spec)
	case "circuit-breaker":
	case "bulkhead":
	case "loadbalancer":
		return nil
	default:
		return fmt.Errorf("not support kind yet")
	}
	return nil
}

func matchValidate(val interface{}) error {
	spec, ok := val.(map[string]interface{})
	if !ok {
		return fmt.Errorf("illegal item : %v", val)
	}
	matches, ok := spec["matches"].([]interface{})
	if !ok {
		return fmt.Errorf("illegal item : %v", spec)
	}
	for _, match := range matches {
		match, ok := match.(map[string]interface{})
		if !ok {
			return fmt.Errorf("illegal item : %v", match)
		}
		if match["name"] == nil {
			return fmt.Errorf("match's name can not be null : %v", match)
		}
		if match["apiPath"] == nil && match["headers"] == nil && match["methods"] == nil {
			return fmt.Errorf("match must have a match item [apiPath/headers/methods] %v", match)
		}
		//apiPath & headers do not check
		if match["methods"] != nil {
			methods, ok := match["methods"].([]interface{})
			if !ok {
				return fmt.Errorf("illegal item : %v", match)
			}
			for _, method := range methods {
				methodStr, ok := method.(string)
				if !ok {
					return fmt.Errorf("illegal item : %v", method)
				}
				if !methodSet[methodStr] {
					return fmt.Errorf("method must be one of the GET/POST/PUT/DELETE: %v", match)
				}
			}
		}
	}
	return nil
}

func retryValidate(val interface{}) error {
	err := policyValidate(val)
	if err != nil {
		return err
	}
	return nil
}

func rateLimitingValidate(val interface{}) error {
	err := policyValidate(val)
	if err != nil {
		return err
	}
	return nil
}

func policyValidate(val interface{}) error {
	//todo : check repeat policy
	spec, ok := val.(map[string]interface{})
	if !ok {
		return fmt.Errorf("illegal item : %v", val)
	}
	rules, ok := spec["rules"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("illegal item : %v", spec)
	}
	if "" == rules["match"] {
		return fmt.Errorf("policy's match can not be nil: %v", spec)
	}
	return nil
}

func init() {
	methodSet = make(map[string]bool)
	methodSet["GET"] = true
	methodSet["POST"] = true
	methodSet["DELETE"] = true
	methodSet["PUT"] = true
}
