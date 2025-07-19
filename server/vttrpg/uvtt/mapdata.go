package uvtt

type MapData struct {
	Format             float64     `json:"format"`
	Image              []byte      `json:"image"`
	Resolution         Resolution  `json:"resolution"`
	LineOfSight        [][]Vector2 `json:"line_of_sight"`
	ObjectsLineOfSight [][]Vector2 `json:"objects_line_of_sight"`
	Portals            []Portal    `json:"portals"`
	Environment        Environment `json:"environment"`
	Lights             []Light     `json:"light"`
}

type Resolution struct {
	// MapOrigin stores the origin meassured in squares
	MapOrigin Vector2 `json:"map_origin"`
	// MapSize stores the size meassured in squares
	MapSize Vector2 `json:"map_size"`
	// PixelsPerGrid stores the amount of pixels per square
	PixelsPerGrid int `json:"pixels_per_grid"`
}

type Portal struct {
	Position     Vector2   `json:"position"`
	Bounds       []Vector2 `json:"bounds"`
	Rotation     float64   `json:"rotation"` // Rotation is meassured in radians
	Closed       bool      `json:"closed"`
	Freestanding bool      `json:"freestanding"`
}

type Environment struct {
	BakedLighting bool   `json:"baked_lighting"`
	AmbientLight  string `json:"ambient_light"` // AmbientLight stores the hex colour code
}

type Light struct {
	Position  Vector2 `json:"position"`
	Range     float64 `json:"range"`
	Intensity float64 `json:"intensity"`
	Color     string  `json:"color"`   // Color stores the hex code of the light
	Shadows   bool    `json:"shadows"` // Whenever or not the light should produce shadows
}

type Vector2 struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}
