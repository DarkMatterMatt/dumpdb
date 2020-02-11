package stringinslice

// StringInSlice checks if `list` contains `s`
func StringInSlice(s string, list []string) bool {
	for _, x := range list {
		if s == x {
			return true
		}
	}
	return false
}
