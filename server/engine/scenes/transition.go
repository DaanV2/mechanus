package scenes

// State identifies what is currently active: a scene and one of its variants.
type State struct {
	SceneID   string `json:"scene_id"`
	VariantID string `json:"variant_id,omitempty"` // Empty means the scene's base instructions
}

// IsZero returns true when no scene is active.
func (s State) IsZero() bool {
	return s.SceneID == ""
}

// Transition describes moving from one state to another. Switching variants within a scene and
// switching scenes are the same operation.
type Transition struct {
	From         State        `json:"from"`
	To           State        `json:"to"`
	Instructions Instructions `json:"instructions"` // The full effective instructions of the target state
	Changes      Instructions `json:"changes"`      // Only what needs to be sent to devices
	Released     []TargetRef  `json:"released"`     // Targets the new state no longer instructs
}

// Plan computes the transition from the current scene/variant to the next one. current may be nil
// when nothing is active yet, in which case every instruction is a change. An empty variant ID
// uses the scene's default variant.
func Plan(current *Scene, currentVariant string, next *Scene, nextVariant string) (*Transition, error) {
	t := &Transition{}

	var prev Instructions
	if current != nil {
		id, err := current.ResolveVariantID(currentVariant)
		if err != nil {
			return nil, err
		}
		prev, err = current.Resolve(id)
		if err != nil {
			return nil, err
		}
		t.From = State{SceneID: current.ID, VariantID: id}
	}

	id, err := next.ResolveVariantID(nextVariant)
	if err != nil {
		return nil, err
	}
	nextIns, err := next.Resolve(id)
	if err != nil {
		return nil, err
	}

	t.To = State{SceneID: next.ID, VariantID: id}
	t.Instructions = nextIns
	t.Changes = prev.Changes(nextIns)
	t.Released = prev.Released(nextIns)

	return t, nil
}
