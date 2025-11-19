package roles

import (
	"errors"
	"slices"
	"strings"
)

// RoleContainer is an interface for types that contain roles.
type RoleContainer interface {
	GetRoles() []string
	SetRoles(roles ...string)
}

// RoleService provides operations for managing roles.
type RoleService struct {
}

// ParseRole parses a string into a Role type, returning an error if invalid.
func ParseRole(role string) (Role, error) {
	v := Role(strings.ToLower(role))
	switch v {
	case Admin, Operator, User, Viewer, Device:
		return v, nil
	default:
		return Viewer, errors.New("unknown role: " + role)
	}
}

// GrantRole grants a role to the container if not already present.
func (r *RoleService) GrantRole(grants RoleContainer, role Role) {
	roles := slices.Clone(grants.GetRoles())
	if slices.Contains(roles, role.String()) {
		return // already granted
	}
	roles = append(roles, role.String())
	grants.SetRoles(roles...)
}

// HasRole checks if the grants contains a role that has the given role, or inherits it.
func (r *RoleService) HasRole(grants RoleContainer, role Role) bool {
	return GrantsHasRole(grants, role)
}

// GrantsHasRole checks if the grants contains a role that has the given role or inherits it.
func GrantsHasRole(grants RoleContainer, role Role) bool {
	for _, r := range grants.GetRoles() {
		if Role(r).Inherits(role) {
			return true
		}
	}

	return false
}
