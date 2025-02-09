package roles

import (
	"testing"
)

func TestRole_Level(t *testing.T) {
	tests := []struct {
		role     Role
		expected int
	}{
		{USER, 0},
		{DEVICE, 0},
		{OPERATOR, 1},
		{ADMIN, 2},
		{"unknown", -1},
	}

	for _, test := range tests {
		if got := test.role.Level(); got != test.expected {
			t.Errorf("Role.Level() = %d, want %d", got, test.expected)
		}
	}
}

func TestRole_Inherits(t *testing.T) {
	tests := []struct {
		role     Role
		inherits Role
		expected bool
	}{
		{USER, DEVICE, true},
		{DEVICE, USER, true},
		{OPERATOR, USER, true},
		{ADMIN, OPERATOR, true},
		{OPERATOR, ADMIN, false},
		{ADMIN, ADMIN, false},
	}

	for _, test := range tests {
		if got := test.role.Inherits(test.inherits); got != test.expected {
			t.Errorf("Role.Inherits(%s) = %v, want %v", test.inherits, got, test.expected)
		}
	}
}

func TestRole_String(t *testing.T) {
	tests := []struct {
		role     Role
		expected string
	}{
		{USER, "user"},
		{DEVICE, "device"},
		{OPERATOR, "operator"},
		{ADMIN, "admin"},
	}

	for _, test := range tests {
		if got := test.role.String(); got != test.expected {
			t.Errorf("Role.String() = %s, want %s", got, test.expected)
		}
	}
}