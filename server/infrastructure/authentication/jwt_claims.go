package authentication

import (
	"slices"

	"github.com/DaanV2/mechanus/server/infrastructure/persistence/models"
	"github.com/golang-jwt/jwt/v5"
)

const (
	// GrantPassword marks a token obtained by logging in with a username and password.
	GrantPassword = "password"
	// GrantRefresh marks a token obtained by refreshing an existing token.
	GrantRefresh = "refresh"

	// ScopeAll allows the token to be used for everything the user's roles allow.
	ScopeAll = "*"
	// ScopeRefresh only allows the token to be used to request a new token.
	ScopeRefresh = "refresh"
)

type (
	// JWTClaims contains the claims for a JWT token.
	JWTClaims struct {
		jwt.RegisteredClaims `json:",inline"`
		User                 JWTUser `json:"user"`
		// The grant under which the token was issued.
		// eg: [GrantPassword], [GrantRefresh]
		Grant string `json:"grant"`
		// What the token may be used for.
		// eg: [ScopeAll], [ScopeRefresh]
		Scope string `json:"scope"`
	}

	// JWTUser contains user information embedded in JWT claims.
	JWTUser struct {
		ID        string   `json:"id"`
		Name      string   `json:"name"`
		Roles     []string `json:"roles"`
		Campaigns []string `json:"campaigns"`
	}
)

// HasUser checks if the claims match the given user.
func (c *JWTClaims) HasUser(user *models.User) bool {
	return c.User.ID == user.ID
}

// HasScope checks if the claims allow the specified scope. [ScopeAll] allows every scope.
func (c *JWTClaims) HasScope(scope string) bool {
	return c.Scope == ScopeAll || c.Scope == scope
}

// GetRoles returns the user's roles from the claims.
func (c *JWTClaims) GetRoles() []string {
	return c.User.Roles
}

// SetRoles sets the user's roles in the claims.
func (c *JWTClaims) SetRoles(newroles ...string) {
	c.User.Roles = newroles
}

// HasCampaign checks if the user has access to the specified campaign.
func (u *JWTUser) HasCampaign(campaign string) bool {
	return slices.Contains(u.Campaigns, campaign)
}

// HasAnyCampaign checks if the user has access to any of the specified campaigns.
func (u *JWTUser) HasAnyCampaign(campaigns ...string) bool {
	return slices.ContainsFunc(campaigns, u.HasCampaign)
}
