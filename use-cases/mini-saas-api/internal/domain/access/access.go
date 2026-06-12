// Package access owns the simplified RBAC role lattice for mini-saas-api.
// It is dependency-free so every domain package and handler can import it.
package access

// Role is a tenant-scoped membership role.
// The lattice is strictly ordered: owner > admin > member.
type Role string

const (
	RoleOwner  Role = "owner"
	RoleAdmin  Role = "admin"
	RoleMember Role = "member"
)

var rank = map[Role]int{
	RoleMember: 1,
	RoleAdmin:  2,
	RoleOwner:  3,
}

// Valid reports whether r is a recognized role.
func (r Role) Valid() bool {
	_, ok := rank[r]
	return ok
}

// AtLeast reports whether r grants at least the privileges of min.
// Unknown roles never satisfy any requirement (fail closed).
func (r Role) AtLeast(min Role) bool {
	rr, ok1 := rank[r]
	mr, ok2 := rank[min]
	return ok1 && ok2 && rr >= mr
}
