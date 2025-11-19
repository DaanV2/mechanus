package xstrings

// FirstNotEmpty returns the first non-empty string from the provided strings, or an empty string if all are empty.
func FirstNotEmpty(items ...string) string {
	for _, item := range items {
		if item != "" {
			return item
		}
	}

	return ""
}
