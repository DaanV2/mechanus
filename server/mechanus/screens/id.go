package screens

import "strings"

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

// Role returns the role portion of the ScreenID.
// Panics if the ScreenID format is invalid (missing ':' separator).
func (s ScreenID) Role() string {
	i := strings.Index(string(s), ":")
	if i == -1 {
		// Note: this should not happen if the ID is created correctly
		panic(ErrInvalidScreenIDFormat)
	}

	return string(s)[:i]
}

// ID returns the identifier portion of the ScreenID.
// Panics if the ScreenID format is invalid (missing ':' separator).
func (s ScreenID) ID() string {
	i := strings.Index(string(s), ":")
	if i == -1 {
		// Note: this should not happen if the ID is created correctly
		panic(ErrInvalidScreenIDFormat)
	}

	return string(s)[i+1:]
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
