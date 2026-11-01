package state

// GridSettings defines the grid layout and positioning for a screen.
type GridSettings struct {
	XOffset   int     `json:"x_offset"`   // Horizontal offset in pixels
	YOffset   int     `json:"y_offset"`   // Vertical offset in pixels
	XRatio    int     `json:"x_ratio"`    // Horizontal ratio for scaling
	YRatio    int     `json:"y_ratio"`    // Vertical ratio for scaling
	Show      bool    `json:"show"`       // Whether the grid lines are drawn
	CellSize  float64 `json:"cell_size"`  // Size of one grid cell in pixels of the view's asset
	LineColor string  `json:"line_color"` // Hex color code of the grid lines
}
