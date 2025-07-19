package screens

import (
	"strings"
)

var (
	ErrInvalidScreenIDFormat = "invalid ScreenID format, missing role separator ':'"
)

// ScreenID represents a unique identifier for a screen, combining a role and an ID.
// The format is "role:id" where role defines the screen's purpose and id is its unique identifier.
type ScreenID string

// NewScreenID creates a new ScreenID by combining a role and ID with a colon separator.
// The resulting format will be "role:id".
func NewScreenID(role, id string) ScreenID {
	return ScreenID(role + ":" + id)
}

// Info returns the role and ID from the ScreenID.
// It splits the ScreenID at the first ':' character.
func (s ScreenID) Info() (role, id string) {
	role, id, found := strings.Cut(string(s), ":")
	if !found {
		panic(ErrInvalidScreenIDFormat)
	}

	return role, id
}

// Role returns the role portion of the ScreenID.
// Panics if the ScreenID format is invalid (missing ':' separator).
func (s ScreenID) Role() string {
	role, _ := s.Info()

	return role
}

// ID returns the identifier portion of the ScreenID.
// Panics if the ScreenID format is invalid (missing ':' separator).
func (s ScreenID) ID() string {
	_, id := s.Info()

	return id
}

// HasRole checks if the ScreenID has the specified role.
// Returns true if the role matches exactly.
func (s ScreenID) HasRole(role string) bool {
	return s.Role() == role
}

// HasID checks if the ScreenID has the specified ID.
// Returns true if the ID matches exactly.
func (s ScreenID) HasID(id string) bool {
	return s.ID() == id
}

// String returns the string representation of the ScreenID in the format "role:id".
// Implements fmt.Stringer interface.
func (s ScreenID) String() string {
	return string(s)
}

// Equals compares this ScreenID with another ScreenID.
// Returns true if both ScreenIDs are exactly equal.
func (s ScreenID) Equals(other ScreenID) bool {
	return s == other
}

// IsEmpty checks if the ScreenID is empty.
// Returns true if the ScreenID is an empty string.
func (s ScreenID) IsEmpty() bool {
	return s == ""
}
