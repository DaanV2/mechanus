package bools

func And(values ...bool) bool {
	for _, v := range values {
		if !v {
			return false
		}
	}

	return true
}

func Or(values ...bool) bool {
	for _, v := range values {
		if v {
			return true
		}
	}

	return false
}
