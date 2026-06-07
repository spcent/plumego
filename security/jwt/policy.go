package jwt

import "strings"

// AuthZPolicy defines role/permission requirements for PolicyAuthorizer.
//
// An empty policy with AllowEmpty false (the zero value) denies all requests.
// Set AllowEmpty true to allow requests when no requirements are configured.
//
// Role and permission comparisons are case-insensitive: "Admin" and "admin"
// are treated as the same value. Callers must normalise role and permission
// strings consistently at issuance time if strict casing is required.
type AuthZPolicy struct {
	// AnyRole requires the principal to hold at least one of these roles.
	AnyRole []string
	// AllRoles requires the principal to hold every one of these roles.
	AllRoles []string
	// AnyPermission requires the principal to hold at least one of these permissions.
	AnyPermission []string
	// AllPermissions requires the principal to hold every one of these permissions.
	AllPermissions []string
	// AllowEmpty permits access when no role or permission requirements are set.
	AllowEmpty bool
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

// contains reports whether target appears in list using case-insensitive comparison.
func contains(list []string, target string) bool {
	for _, v := range list {
		if strings.EqualFold(v, target) {
			return true
		}
	}
	return false
}
