package jwt

import "strings"

// AuthZPolicy defines role/permission requirements.
type AuthZPolicy struct {
	AnyRole        []string
	AllRoles       []string
	AnyPermission  []string
	AllPermissions []string
	AllowEmpty     bool
}

func checkPolicy(policy AuthZPolicy, auth AuthorizationClaims) bool {
	if !policy.AllowEmpty && !policyHasRequirements(policy) {
		return false
	}

	hasAll := func(required []string, actual []string) bool {
		for _, r := range required {
			if !contains(actual, r) {
				return false
			}
		}
		return true
	}
	hasAny := func(required []string, actual []string) bool {
		if len(required) == 0 {
			return true
		}
		for _, r := range required {
			if contains(actual, r) {
				return true
			}
		}
		return false
	}

	if len(policy.AllRoles) > 0 && !hasAll(policy.AllRoles, auth.Roles) {
		return false
	}
	if len(policy.AllPermissions) > 0 && !hasAll(policy.AllPermissions, auth.Permissions) {
		return false
	}
	if !hasAny(policy.AnyRole, auth.Roles) {
		return false
	}
	if !hasAny(policy.AnyPermission, auth.Permissions) {
		return false
	}
	return true
}

func policyHasRequirements(policy AuthZPolicy) bool {
	return len(policy.AnyRole) > 0 ||
		len(policy.AllRoles) > 0 ||
		len(policy.AnyPermission) > 0 ||
		len(policy.AllPermissions) > 0
}

func contains(list []string, target string) bool {
	for _, v := range list {
		if strings.EqualFold(v, target) {
			return true
		}
	}
	return false
}
