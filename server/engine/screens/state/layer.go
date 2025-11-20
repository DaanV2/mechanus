package state

// Layer represents a rendering layer containing entities.
type Layer struct {
	Entities map[string]*Entity `json:"entities"` // Map of entity IDs to Entity objects
}
