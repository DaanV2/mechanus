package state

import "github.com/google/uuid"

// Asset represents a visual or audio asset used in a screen.
type Asset struct {
	ID        string `json:"id"`                   // Unique identifier for the asset
	Name      string `json:"name"`                 // Display name of the asset
	Url       string `json:"url"`                  // Where the client loads the asset from
	MediaType string `json:"media_type,omitempty"` // MIME type, e.g. "image/png", empty when unknown
}

// NewAssetFrom creates a new asset from a URL.
func NewAssetFrom(url string) *Asset {
	return &Asset{
		ID:  uuid.NewString(),
		Url: url,
	}
}
