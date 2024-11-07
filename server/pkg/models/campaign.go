package models

type Campaign struct {
	BaseItem `json:",inline"`

	Players []string `json:"players"`
}

func (c Campaign) GetPlayers() []string { return c.Players }