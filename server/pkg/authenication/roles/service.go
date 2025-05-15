package roles

import (
	"errors"
	"slices"
	"strings"
)

const (
	Admin    Role = "admin"
	Operator Role = "operator"
	User     Role = "user"
	Viewer   Role = "viewer"
)

type Role string

func (r Role) String() string {
	return string(r)
}

func (r Role) Inherits(other Role) bool {
	return r.value() >= other.value()
}

func (r Role) value() int {
	switch r {
	case Admin:
		return 3
	case Operator:
		return 2
	case User:
		return 1
	case Viewer:
		return 0
	}

	return -1
}

type RoleContainer interface {
	GetRoles() []string
	SetRoles(roles ...string)
}

type RoleService struct {
}

func ParseRole(role string) (Role, error) {
	v := Role(strings.ToLower(role))
	switch v {
	case Admin, Operator, User, Viewer:
		return v, nil
	default:
		return Viewer, errors.New("unknown role: " + role)
	}
}

func (r *RoleService) GrantRole(container RoleContainer, role Role) {
	roles := slices.Clone(container.GetRoles())
	roles = append(roles, role.String())
	container.SetRoles(roles...)
}

func (r *RoleService) HasRole(container RoleContainer, role Role) bool {
	for _, r := range container.GetRoles() {
		if role.Inherits(Role(r)) {
			return true
		}
	}

	return false
}
