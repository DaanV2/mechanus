package state

import "github.com/google/uuid"

type Asset struct {
	ID  string `json:"id"`   // Unique identifier for the asset
	Url string `json:"name"` // Name of the asset
}

func NewAssetFrom(url string) *Asset {
	return &Asset{
		ID:  uuid.NewString(),
		Url: url,
	}
}
