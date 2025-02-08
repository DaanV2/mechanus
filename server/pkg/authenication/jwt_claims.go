package authenication

import (
	"slices"

	"github.com/DaanV2/mechanus/server/pkg/models/users"
	"github.com/golang-jwt/jwt/v5"
)

type (
	JWTClaims struct {
		jwt.RegisteredClaims `json:",inline"`
		User                 JWTUser `json:"user"`
		Scope                string  `json:"scope"`
	}

	JWTUser struct {
		ID        string   `json:"id"`
		Name      string   `json:"name"`
		Roles     []string `json:"roles"`
		Campaigns []string `json:"campaigns"`
	}
)

func (c *JWTClaims) HasUser(user *users.User) bool {
	return c.User.ID == user.ID
}

func (c *JWTClaims) HasScope(scope string) bool {
	return c.Scope == scope
}

func (u *JWTUser) HasRole(role string) bool {
	return slices.Contains(u.Roles, role)
}

func (u *JWTUser) HasAnyRole(roles ...string) bool {
	for _, r := range roles {
		if u.HasRole(r) {
			return true
		}
	}

	return false
}

func (u *JWTUser) HasCampaign(campaign string) bool {
	return slices.Contains(u.Campaigns, campaign)
}

func (u *JWTUser) HasAnyCampaign(campaigns ...string) bool {
	for _, c := range campaigns {
		if u.HasCampaign(c) {
			return true
		}
	}

	return false
}
