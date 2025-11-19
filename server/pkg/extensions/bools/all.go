package bools

// And returns true if all values are true. Returns true for an empty slice.
func And(values ...bool) bool {
	for _, v := range values {
		if !v {
			return false
		}
	}

	return true
}

// Or returns true if any value is true. Returns false for an empty slice.
func Or(values ...bool) bool {
	for _, v := range values {
		if v {
			return true
		}
	}

	return false
}
