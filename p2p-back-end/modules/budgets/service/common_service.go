package service

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

func getColSafe(row []string, idx int) string {
	if idx < len(row) {
		return row[idx]
	}
	return ""
}

func parseDecimal(s string) decimal.Decimal {
	if s == "" {
		return decimal.Zero
	}
	// Robust Parsing: Remove commas and spaces
	cleanS := strings.ReplaceAll(s, ",", "")
	cleanS = strings.TrimSpace(cleanS)

	d, err := decimal.NewFromString(cleanS)
	if err != nil {
		// Log warning only for non-empty distinct strings to avoid spam
		if len(cleanS) > 0 {
			fmt.Printf("[Parse Warning] Invalid decimal: '%s' -> 0\n", s)
		}
		return decimal.Zero
	}
	return d
}

func extractCode(s string) string {
	if strings.Contains(s, " - ") {
		parts := strings.SplitN(s, " - ", 2)
		return strings.TrimSpace(parts[0])
	}
	return s
}

func sanitizeFilter(filter map[string]interface{}) {

	// Normalize keys: Ensure entities/branches exist provided entity/branch exist
	if v, ok := filter["entity"]; ok {
		if _, exists := filter["entities"]; !exists {
			filter["entities"] = v
		}
	}
	if v, ok := filter["branch"]; ok {
		if _, exists := filter["branches"]; !exists {
			filter["branches"] = v
		}
	}
	// Normalize keys: Department
	if v, ok := filter["department"]; ok {
		if _, exists := filter["departments"]; !exists {
			filter["departments"] = v
		}
	}

	targetKeys := []string{"entities", "branches", "departments"}
	for _, key := range targetKeys {
		val, ok := filter[key]
		if !ok {
			continue
		}

		var finalSlice []string

		// Case 1: Single String
		if s, ok := val.(string); ok && s != "" {
			finalSlice = append(finalSlice, extractCode(s))
		} else if ss, ok := val.([]string); ok {
			// Case 2: Slice of Strings
			for _, v := range ss {
				finalSlice = append(finalSlice, extractCode(v))
			}
		} else if ifaceSlice, ok := val.([]interface{}); ok {
			// Case 3: Slice of Interface
			for _, v := range ifaceSlice {
				if s, ok := v.(string); ok {
					finalSlice = append(finalSlice, extractCode(s))
				}
			}
		}

		// Update Filter with correct type []string
		if len(finalSlice) > 0 {
			filter[key] = finalSlice
		} else {
			// If empty, remove to avoid confusion or empty IN clause issues
			delete(filter, key)
		}
	}
}




func extractYear(s string) string {
	// Simple scan for 4 digits starting with 20
	// e.g. "Budget 2025" or "FY2025"
	for i := 0; i <= len(s)-4; i++ {
		sub := s[i : i+4]
		// Check if it's a number and starts with "20"
		if sub >= "2010" && sub <= "2099" {
			if isNumeric(sub) {
				return sub
			}
		}
	}
	return ""
}

func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
