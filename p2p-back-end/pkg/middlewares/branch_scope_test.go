package middlewares

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"p2p-back-end/modules/entities/models"
)

func TestEnforceBranchScope_NoUser_NoOp(t *testing.T) {
	filter := map[string]interface{}{"branches": []string{"HOF"}}
	EnforceBranchScope(nil, filter)
	assert.Equal(t, []string{"HOF"}, filter["branches"],
		"missing user should be a no-op (caller-provided filter preserved)")
}

func TestEnforceBranchScope_NotBranchDelegate_NoOp(t *testing.T) {
	user := &models.UserInfo{
		ID:          "u-1",
		Roles:       []string{"DELEGATE"}, // plain DELEGATE, NOT BRANCH_DELEGATE
		BranchCodes: []string{"HOF"},
	}
	filter := map[string]interface{}{"branches": []string{"SUR"}}
	EnforceBranchScope(user, filter)
	assert.Equal(t, []string{"SUR"}, filter["branches"],
		"plain DELEGATE must NOT have branch scope enforced")
}

func TestEnforceBranchScope_AdminUnaffected(t *testing.T) {
	user := &models.UserInfo{
		ID:          "u-2",
		Roles:       []string{"ADMIN"},
		BranchCodes: []string{"HOF"},
	}
	filter := map[string]interface{}{"branches": []string{"SUR", "BUR"}}
	EnforceBranchScope(user, filter)
	assert.Equal(t, []string{"SUR", "BUR"}, filter["branches"],
		"ADMIN must see anything they ask for, scope not enforced")
}

func TestEnforceBranchScope_BranchDelegate_SingleCode_Override(t *testing.T) {
	user := &models.UserInfo{
		ID:          "u-3",
		Roles:       []string{"BRANCH_DELEGATE"},
		BranchCodes: []string{"SUR"},
	}
	// Caller tries to widen scope by sending many branches
	filter := map[string]interface{}{"branches": []string{"HOF", "BUR", "Branch01"}}
	EnforceBranchScope(user, filter)
	assert.Equal(t, []string{"SUR"}, filter["branches"],
		"BRANCH_DELEGATE filter must be overridden to user's own branch only")
}

func TestEnforceBranchScope_BranchDelegate_MultipleCodes_AllInjected(t *testing.T) {
	// CLIK HQ case — same logical branch with two code variants in actual data
	user := &models.UserInfo{
		ID:          "u-clik-hq",
		Roles:       []string{"BRANCH_DELEGATE"},
		BranchCodes: []string{"HOF", "Branch00"},
	}
	filter := map[string]interface{}{}
	EnforceBranchScope(user, filter)
	assert.ElementsMatch(t, []string{"HOF", "Branch00"}, filter["branches"],
		"all codes mapped to the user's company must end up in branches filter")
}

func TestEnforceBranchScope_BranchDelegate_MultipleCodes_OverridesCallerFilter(t *testing.T) {
	user := &models.UserInfo{
		ID:          "u-clik-hq",
		Roles:       []string{"BRANCH_DELEGATE"},
		BranchCodes: []string{"HOF", "Branch00"},
	}
	// Caller tries to widen — must be replaced by user's own codes only
	filter := map[string]interface{}{"branches": []string{"SUR", "BUR"}}
	EnforceBranchScope(user, filter)
	assert.ElementsMatch(t, []string{"HOF", "Branch00"}, filter["branches"],
		"caller filter must be REPLACED, not merged")
}

func TestEnforceBranchScope_BranchDelegate_InjectsWhenAbsent(t *testing.T) {
	user := &models.UserInfo{
		ID:          "u-4",
		Roles:       []string{"BRANCH_DELEGATE"},
		BranchCodes: []string{"Branch07"},
	}
	filter := map[string]interface{}{} // caller sent no branches
	EnforceBranchScope(user, filter)
	assert.Equal(t, []string{"Branch07"}, filter["branches"],
		"BRANCH_DELEGATE must have branch injected even if caller omits it")
}

func TestEnforceBranchScope_BranchDelegate_NoMappingDeniesAll(t *testing.T) {
	user := &models.UserInfo{
		ID:    "u-5",
		Roles: []string{"BRANCH_DELEGATE"},
		// BranchCodes empty — admin hasn't configured a mapping
	}
	filter := map[string]interface{}{"branches": []string{"HOF"}}
	EnforceBranchScope(user, filter)
	assert.Equal(t, []string{"__NO_BRANCH__"}, filter["branches"],
		"BRANCH_DELEGATE without mapping must get sentinel that matches no rows (deny by default)")
}

func TestEnforceBranchScope_NilFilterIsSafe(t *testing.T) {
	user := &models.UserInfo{
		ID:          "u-6",
		Roles:       []string{"BRANCH_DELEGATE"},
		BranchCodes: []string{"HOF"},
	}
	assert.NotPanics(t, func() {
		EnforceBranchScope(user, nil)
	})
}

func TestEnforceBranchScope_RoleMatchIsCaseInsensitive(t *testing.T) {
	user := &models.UserInfo{
		ID:          "u-7",
		Roles:       []string{"branch_delegate"}, // lowercase
		BranchCodes: []string{"MKB"},
	}
	filter := map[string]interface{}{}
	EnforceBranchScope(user, filter)
	assert.Equal(t, []string{"MKB"}, filter["branches"],
		"role check must be case-insensitive (matches existing RolesGuard behavior)")
}

func TestEnforceBranchScope_MultiRoleIncludingBranchDelegate(t *testing.T) {
	user := &models.UserInfo{
		ID:          "u-8",
		Roles:       []string{"OWNER", "BRANCH_DELEGATE"}, // multi-role
		BranchCodes: []string{"AVN"},
	}
	filter := map[string]interface{}{"branches": []string{"HOF"}}
	EnforceBranchScope(user, filter)
	assert.Equal(t, []string{"AVN"}, filter["branches"],
		"BRANCH_DELEGATE scope must be enforced even when paired with another role")
}

func TestEnforceBranchScope_DoesNotMutateUserBranchCodes(t *testing.T) {
	// Defensive: filter mutation must not bleed into UserInfo.BranchCodes
	user := &models.UserInfo{
		ID:          "u-9",
		Roles:       []string{"BRANCH_DELEGATE"},
		BranchCodes: []string{"HOF", "Branch00"},
	}
	filter := map[string]interface{}{}
	EnforceBranchScope(user, filter)

	// Mutate the filter slice — should NOT affect user.BranchCodes
	branches := filter["branches"].([]string)
	branches[0] = "MUTATED"

	assert.Equal(t, []string{"HOF", "Branch00"}, user.BranchCodes,
		"filter slice must be a defensive copy, not aliased to UserInfo.BranchCodes")
}

func TestHasBranchDelegateRole(t *testing.T) {
	cases := []struct {
		name  string
		roles []string
		want  bool
	}{
		{"empty", nil, false},
		{"only admin", []string{"ADMIN"}, false},
		{"only delegate", []string{"DELEGATE"}, false},
		{"branch delegate present", []string{"OWNER", "BRANCH_DELEGATE"}, true},
		{"lowercase variant", []string{"branch_delegate"}, true},
		{"mixed case", []string{"Branch_Delegate"}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, hasBranchDelegateRole(tc.roles))
		})
	}
}
