package state

// Layer represents a rendering layer containing entities.
type Layer struct {
	ID       string             `json:"id"`       // Unique identifier for the layer within the screen
	Entities map[string]*Entity `json:"entities"` // Map of entity IDs to Entity objects
}
