package state

type GridSettings struct {
	XOffset int `json:"x_offset"` // Horizontal offset in pixels
	YOffset int `json:"y_offset"` // Vertical offset in pixels
	XRatio  int `json:"x_ratio"`  // Horizontal ratio for scaling
	YRatio  int `json:"y_ratio"`  // Vertical ratio for scaling
}
