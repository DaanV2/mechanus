package authenication

import (
	"slices"

	"github.com/DaanV2/mechanus/server/pkg/database/models"
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

func (c *JWTClaims) HasUser(user *models.User) bool {
	return c.User.ID == user.ID
}

func (c *JWTClaims) HasScope(scope string) bool {
	return c.Scope == scope
}

func (c *JWTClaims) GetRoles() []string {
	return c.User.Roles
}
func (c *JWTClaims) SetRoles(roles ...string) {
	c.User.Roles = roles
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
