package scenes

// Ref identifies a scene and one of its variants.
type Ref struct {
	SceneID   string `json:"scene_id"`
	VariantID string `json:"variant_id,omitempty"` // Empty means the scene's base instructions
}

// IsZero returns true when no scene is referenced.
func (r Ref) IsZero() bool {
	return r.SceneID == ""
}

// Transition describes what changes for the devices when moving from one set of instructions to
// another: switching scenes, switching variants, and live adjustments are all transitions.
type Transition struct {
	From         Ref          `json:"from"`
	To           Ref          `json:"to"`
	Instructions Instructions `json:"instructions"` // The full effective instructions after the transition
	Changes      Instructions `json:"changes"`      // Only what needs to be sent to devices
	Released     []TargetRef  `json:"released"`     // Targets that are no longer instructed
}

// NewTransition computes the transition from prev to next. Use a zero [Ref] and empty
// [Instructions] for from/prev when nothing was active, in which case every instruction is a change.
func NewTransition(from Ref, prev Instructions, to Ref, next Instructions) *Transition {
	return &Transition{
		From:         from,
		To:           to,
		Instructions: next,
		Changes:      prev.Changes(next),
		Released:     prev.Released(next),
	}
}
