package middlewares

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"p2p-back-end/modules/entities/models"
)

// EnforceBranchScopeFromCtx is the fiber-aware wrapper that pulls UserInfo
// from request Locals and delegates to EnforceBranchScope.
func EnforceBranchScopeFromCtx(c *fiber.Ctx, filter map[string]interface{}) {
	EnforceBranchScope(UserFromCtx(c), filter)
}

// EnforceBranchScope forces filter["branches"] to the user's resolved branch
// codes when the user has the BRANCH_DELEGATE role. It overrides any
// caller-supplied branches filter (so a malicious frontend cannot widen its
// scope) and uses ALL codes mapped to the user's company (1 company → many
// codes via company_branch_code_mappings).
//
// No-op when:
//   - user is nil
//   - user has no BRANCH_DELEGATE role
//   - filter is nil
//
// When user IS BRANCH_DELEGATE but has no resolved BranchCodes (mapping
// missing), "branches" is set to a sentinel that matches nothing, ensuring
// the user sees zero rows rather than all rows.
func EnforceBranchScope(user *models.UserInfo, filter map[string]interface{}) {
	if user == nil || filter == nil {
		return
	}
	if !hasBranchDelegateRole(user.Roles) {
		return
	}

	if len(user.BranchCodes) == 0 {
		filter["branches"] = []string{"__NO_BRANCH__"}
		return
	}
	// Copy to avoid sharing the slice with UserInfo — downstream might mutate.
	codes := make([]string, len(user.BranchCodes))
	copy(codes, user.BranchCodes)
	filter["branches"] = codes
}

// UserFromCtx extracts the authenticated UserInfo previously stored by
// JwtAuthentication into fiber Locals. Returns nil if missing or wrong type.
func UserFromCtx(c *fiber.Ctx) *models.UserInfo {
	v := c.Locals("user")
	if v == nil {
		return nil
	}
	user, ok := v.(*models.UserInfo)
	if !ok {
		return nil
	}
	return user
}

func hasBranchDelegateRole(roles []string) bool {
	for _, r := range roles {
		if strings.EqualFold(r, models.RoleBranchDelegate) {
			return true
		}
	}
	return false
}
