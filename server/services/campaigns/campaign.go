package campaigns

import "github.com/DaanV2/mechanus/server/pkg/database"

type Campaign struct {
	database.BaseItem

	Players []string `json:"players"`
}
