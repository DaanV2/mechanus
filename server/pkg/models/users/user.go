package users

import (
	"github.com/DaanV2/mechanus/server/pkg/models"
	"github.com/DaanV2/mechanus/server/pkg/models/roles"
)

type User struct {
	models.BaseItem `json:",inline"`
	Username        string       `json:"username"`
	Roles           []roles.Role `json:"roles"`
	Campaigns       []string     `json:"campaigns"`
	PasswordHash    []byte       `json:"passwordhash"`
}
