package semver

import "strings"

type Range struct {
	comparators []*comparator
}

func MustParseRange(raw string) *Range {
	r, err := ParseRange(raw)
	must(err)
	return r
}

func ParseRange(raw string) (*Range, error) {
	r := &Range{
		comparators: []*comparator{},
	}
	for _, raw := range strings.Split(raw, " ") {
		c, err := parseComparator(raw)
		if err != nil {
			return nil, err
		}
		r.comparators = append(r.comparators, c)
	}
	return r, nil
}

func (r *Range) Valid(v *Version) bool {
	for _, c := range r.comparators {
		if !c.valid(v) {
			return false
		}
	}
	return true
}
