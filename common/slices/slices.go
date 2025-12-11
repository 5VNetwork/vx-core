package slices

func CompareSlices[T comparable](slice1, slice2 []T) bool {
	if len(slice1) != len(slice2) {
		return false
	}

	// Create maps to store element counts
	count1 := make(map[T]int)
	count2 := make(map[T]int)

	// Count elements in first slice
	for _, elem := range slice1 {
		count1[elem]++
	}

	// Count elements in second slice
	for _, elem := range slice2 {
		count2[elem]++
	}

	// Compare the maps
	if len(count1) != len(count2) {
		return false
	}

	for elem, count := range count1 {
		if count2[elem] != count {
			return false
		}
	}

	return true
}
