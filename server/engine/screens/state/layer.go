package state

type Layer struct {
	Entities map[string]*Entity `json:"entities"` // Map of entity IDs to Entity objects
}
