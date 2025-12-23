package validator

// FilterByCode returns a subset of issues matching any of the given codes.
func FilterByCode(issues []Issue, codes ...string) []Issue {
	set := map[string]struct{}{}
	for _, c := range codes {
		set[c] = struct{}{}
	}
	out := make([]Issue, 0, len(issues))
	for _, is := range issues {
		if _, ok := set[is.Code]; ok {
			out = append(out, is)
		}
	}
	return out
}
