package state

import "github.com/google/uuid"

// Asset represents a visual or audio asset used in a screen.
type Asset struct {
	ID  string `json:"id"`   // Unique identifier for the asset
	Url string `json:"name"` // Name of the asset
}

// NewAssetFrom creates a new asset from a URL.
func NewAssetFrom(url string) *Asset {
	return &Asset{
		ID:  uuid.NewString(),
		Url: url,
	}
}
