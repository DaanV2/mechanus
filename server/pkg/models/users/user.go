package users

import "github.com/DaanV2/mechanus/server/pkg/models"

type User struct {
	models.BaseItem `json:",inline"`
	Name            string   `json:"name"`
	Roles           []string `json:"roles"`
	Campaigns       []string `json:"campaigns"`
	PasswordHash    []byte   `json:"passwordhash"`
}
