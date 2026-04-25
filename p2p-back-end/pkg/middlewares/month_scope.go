package middlewares

import (
	"fmt"
	"strings"

	"p2p-back-end/modules/entities/models"
)

// EnforceMonthScope restricts filter["months"] to the admin's permitted months
// for non-admin users. Admin bypasses this check (sees anything).
//
// Behavior:
//   - user nil / filter nil → no-op
//   - user has ADMIN role → no-op
//   - permittedMonths empty → filter["months"] = sentinel matching nothing
//   - request has no months → filter["months"] = permittedMonths (full scope)
//   - request has months → filter["months"] = intersection(request, permitted)
//   - intersection is empty → sentinel matching nothing
func EnforceMonthScope(user *models.UserInfo, filter map[string]interface{}, permittedMonths []string) {
	if user == nil || filter == nil {
		return
	}
	if hasAdminRole(user.Roles) {
		return
	}

	if len(permittedMonths) == 0 {
		filter["months"] = []string{"__NO_MONTH__"}
		return
	}

	reqMonths := extractMonths(filter["months"])
	if len(reqMonths) == 0 {
		filter["months"] = permittedMonths
		return
	}

	permSet := make(map[string]struct{}, len(permittedMonths))
	for _, m := range permittedMonths {
		permSet[strings.ToUpper(strings.TrimSpace(m))] = struct{}{}
	}

	var intersection []string
	for _, m := range reqMonths {
		if _, ok := permSet[strings.ToUpper(strings.TrimSpace(m))]; ok {
			intersection = append(intersection, m)
		}
	}
	if len(intersection) == 0 {
		filter["months"] = []string{"__NO_MONTH__"}
		return
	}
	filter["months"] = intersection
}

func hasAdminRole(roles []string) bool {
	for _, r := range roles {
		if strings.Contains(strings.ToUpper(r), "ADMIN") {
			return true
		}
	}
	return false
}

func extractMonths(val interface{}) []string {
	if val == nil {
		return nil
	}
	switch v := val.(type) {
	case []string:
		return v
	case []interface{}:
		var out []string
		for _, item := range v {
			if s, ok := item.(string); ok && s != "" {
				out = append(out, s)
			} else if item != nil {
				out = append(out, fmt.Sprintf("%v", item))
			}
		}
		return out
	case string:
		if v == "" {
			return nil
		}
		return []string{v}
	}
	return nil
}
