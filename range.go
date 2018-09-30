package semver

import (
	"bytes"
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
)

type comparators []*comparator
type comparatorSet []comparators
type Range struct {
	set comparatorSet
}

func MustParseRange(raw string) *Range {
	r, err := ParseRange(raw)
	must(err)
	return r
}

func atoi(s string) int {
	i, err := strconv.Atoi(s)
	must(err)
	return i
}
func itoa(i int) string {
	return strconv.Itoa(i)
}

func isX(id string) bool {
	return id == "" || strings.ToLower(id) == "x" || id == "*"
}

func ParseRange(raw string) (*Range, error) {
	raw = reHyphenRange.ReplaceAllStringFunc(raw, func(raw string) string {
		submatches := reHyphenRange.FindStringSubmatch(raw)
		from := submatches[1]
		fM := submatches[2]
		fm := submatches[3]
		fp := submatches[4]
		// fpr := submatches[5]
		// fb := submatches[6]
		to := submatches[7]
		tM := submatches[8]
		tm := submatches[9]
		tp := submatches[10]
		tpr := submatches[11]
		// tb := submatches[12]
		if isX(fM) {
			from = ""
		} else if isX(fm) {
			from = ">=" + fM + ".0.0"
		} else if isX(fp) {
			from = ">=" + fM + "." + fm + ".0"
		} else {
			from = ">=" + from
		}
		if isX(tM) {
			to = ""
		} else if isX(tm) {
			to = "<" + itoa((atoi(tM) + 1)) + ".0.0"
		} else if isX(tp) {
			to = "<" + tM + "." + itoa(atoi(tm)+1) + ".0"
		} else if tpr != "" {
			to = "<=" + tM + "." + tm + "." + tp + "-" + tpr
		} else {
			to = "<=" + to
		}
		return strings.TrimSpace(from + " " + to)
	})
	raw = reComparatorTrim.ReplaceAllString(raw, "$1$2$3")
	raw = reTildeTrim.ReplaceAllString(raw, "$1~")
	raw = reCaretTrim.ReplaceAllString(raw, "$1^")
	raw = strings.Join(regexp.MustCompile(`\s+`).Split(raw, -1), " ")
	raw = replaceTildes(raw)
	raw = replaceCarets(raw)
	raw = replaceXRanges(raw)
	raw = replaceStars(raw)
	r := &Range{
		set: comparatorSet{},
	}
	for _, raw := range regexp.MustCompile(`\s*\|\|\s*`).Split(raw, -1) {
		comparators := []*comparator{}
		for _, raw := range strings.Split(raw, " ") {
			c, err := parseComparator(raw)
			if err != nil {
				return nil, err
			}
			comparators = append(comparators, c)
		}
		r.set = append(r.set, comparators)
	}
	return r, nil
}

func (r *Range) Valid(v *Version) bool {
	return r.set.Valid(v)
}

func (r *Range) String() string {
	var i []string
	for _, comparators := range r.set {
		var j []string
		for _, c := range comparators {
			j = append(j, c.String())
		}
		i = append(i, strings.Join(j, " "))
	}
	return strings.Join(i, " || ")
}

func (set comparatorSet) Valid(v *Version) bool {
	for _, comparators := range set {
		if comparators.Valid(v) {
			return true
		}
	}
	return false
}

func (comparators comparators) Valid(v *Version) bool {
	for _, c := range comparators {
		if !c.valid(v) {
			return false
		}
	}
	return true
}

func (this *Range) MarshalJSON() ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(this.String())
	return buffer.Bytes(), err
}

func (this *Range) UnmarshalJSON(b []byte) error {
	v, err := ParseRange(string(b))
	if err != nil {
		return err
	}
	this.set = v.set
	return nil
}

// func (this *Range) GT(v *Version) bool {
// 	return true
// }
