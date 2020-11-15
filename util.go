package wilddawg

func sameMachineEdges(a map[interface{}]StateId,
	b map[interface{}]StateId) bool {
	if len(a) != len(b) {
		return false
	}
	for k, a_val := range a {
		if b_val, present := b[k]; !present {
			return false
		} else if a_val != b_val {
			return false
		}
	}
	return true
}

func slicesSameValues(a []interface{}, b []interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	count := make(map[interface{}]int)
	for _, el := range a {
		count[el] += 1
	}
	for _, el := range b {
		count[el] -= 1
	}
	for _, c := range count {
		if c != 0 {
			return false
		}
	}
	return true
}
