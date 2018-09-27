package semver

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

var numericIdentifier = `0|[1-9]\d*`
var mainVersion = `(` + numericIdentifier + `)\.(` + numericIdentifier + `)\.(` + numericIdentifier + `)`

var reNumericIdentifier = regexp.MustCompile(numericIdentifier)
var reMainVersion = regexp.MustCompile(mainVersion)

type Version struct {
	Major int
	Minor int
	Patch int
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func MustParse(raw string) *Version {
	v, err := Parse(raw)
	must(err)
	return v
}

func Parse(raw string) (*Version, error) {
	parts := make([]int, 3)
	submatches := reMainVersion.FindStringSubmatch(raw)
	if len(submatches) == 0 {
		return nil, errors.New("invalid version: " + raw)
	}
	for i, s := range submatches[1:] {
		var err error
		parts[i], err = strconv.Atoi(s)
		if err != nil {
			return nil, err
		}
	}
	return &Version{
		Major: parts[0],
		Minor: parts[1],
		Patch: parts[2],
	}, nil
}

func (this *Version) String() string {
	return fmt.Sprintf("%d.%d.%d", this.Major, this.Minor, this.Patch)
}

func (this *Version) MarshalJSON() ([]byte, error) {
	return []byte(`"` + this.String() + `"`), nil
}

func (this *Version) UnmarshalJSON(b []byte) error {
	v, err := Parse(string(b))
	if err != nil {
		return err
	}
	this.Major = v.Major
	this.Minor = v.Minor
	this.Patch = v.Patch
	return nil
}

func compare(a, b int) int {
	if a < b {
		return -1
	}
	if b < a {
		return 1
	}
	return 0
}

func (a *Version) compareMajor(b *Version) int {
	return compare(a.Major, b.Major)
}

func (a *Version) compareMinor(b *Version) int {
	return compare(a.Minor, b.Minor)
}

func (a *Version) comparePatch(b *Version) int {
	return compare(a.Patch, b.Patch)
}
func (a *Version) compare(b *Version) int {
	var c int
	c = a.compareMajor(b)
	if c != 0 {
		return c
	}
	c = a.compareMinor(b)
	if c != 0 {
		return c
	}
	return a.comparePatch(b)
}

// LT returns true is given version is less than this one
func (a *Version) LT(b *Version) bool {
	return a.compare(b) == -1
}

type Versions []*Version

func (v Versions) Len() int {
	return len(v)
}
func (v Versions) Less(a, b int) bool {
	return v[a].LT(v[b])
}
func (v Versions) Swap(a, b int) {
	v[a], v[b] = v[b], v[a]
}
