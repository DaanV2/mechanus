package authentication

import (
	"slices"

	"github.com/DaanV2/mechanus/server/infrastructure/persistence/models"
	"github.com/golang-jwt/jwt/v5"
)

type (
	// JWTClaims contains the claims for a JWT token.
	JWTClaims struct {
		jwt.RegisteredClaims `json:",inline"`
		User                 JWTUser `json:"user"`
		// The scope under which the token was issued.
		// eg: "password", "refresh"
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

// HasScope checks if the claims have the specified scope.
func (c *JWTClaims) HasScope(scope string) bool {
	return c.Scope == scope
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
	for _, c := range campaigns {
		if u.HasCampaign(c) {
			return true
		}
	}

	return false
}
