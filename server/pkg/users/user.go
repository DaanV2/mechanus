package users

import "github.com/DaanV2/mechanus/server/pkg/database"

type User struct {
	database.BaseItem
	Name         string   `json:"name"`
	Roles        []string `json:"roles"`
	Campaigns    []string `json:"campaigns"`
	PasswordHash []byte   `json:"passwordhash"`
}
