package utils

import (
	"strings"
)

// SanitizeFilter cleans and normalizes filter keys and values
func SanitizeFilter(key string, val interface{}) (string, []string) {
	// 1. Normalize Keys (e.g., 'entity' -> 'entities')
	normalizedKey := key
	switch strings.ToLower(key) {
	case "entity":
		normalizedKey = "entities"
	case "branch":
		normalizedKey = "branches"
	case "department":
		normalizedKey = "departments"
	case "conso_gl":
		normalizedKey = "conso_gls"
	}

	// 2. Normalize Values (Ensure []string and strip " - " labels)
	var result []string
	switch v := val.(type) {
	case string:
		if v != "" {
			result = []string{stripLabel(v)}
		}
	case []interface{}:
		for _, item := range v {
			if s, ok := item.(string); ok && s != "" {
				result = append(result, stripLabel(s))
			}
		}
	case []string:
		for _, s := range v {
			if s != "" {
				result = append(result, stripLabel(s))
			}
		}
	}

	return normalizedKey, result
}

// stripLabel removes " - Company Name" suffix from "HMW - HMW Company"
func stripLabel(s string) string {
	if idx := strings.Index(s, " - "); idx != -1 {
		return strings.TrimSpace(s[:idx])
	}
	return strings.TrimSpace(s)
}

// toStringSlice converts a value to a string slice safely
func ToStringSlice(v interface{}) []string {
	if v == nil {
		return nil
	}
	switch val := v.(type) {
	case string:
		if val == "" {
			return nil
		}
		return []string{val}
	case []string:
		return val
	case []interface{}:
		var res []string
		for _, item := range val {
			if s, ok := item.(string); ok && s != "" {
				res = append(res, s)
			}
		}
		return res
	default:
		return nil
	}
}

// GetSafeString retrieves a string from a map safely, handling both string and []string types
func GetSafeString(data map[string]interface{}, key string) string {
	if val, ok := data[key]; ok {
		if s, ok := val.(string); ok {
			return s
		}
		if arr, ok := val.([]string); ok && len(arr) > 0 {
			return arr[0]
		}
	}
	return ""
}

// ConvertInterfaceSliceToStringSlice handles []interface{} to []string conversion
func ConvertInterfaceSliceToStringSlice(v []interface{}) []string {
	var res []string
	for _, item := range v {
		if s, ok := item.(string); ok {
			res = append(res, s)
		}
	}
	return res
}
