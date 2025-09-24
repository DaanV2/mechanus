package xstrings

func FirstNotEmpty(items ...string) string {
	for _, item := range items {
		if item != "" {
			return item
		}
	}

	return ""
}
