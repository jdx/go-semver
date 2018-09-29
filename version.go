package semver

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var numericIdentifier = `0|[1-9]\d*`
var nonNumericIdentifier = `\d*[a-zA-Z-][a-zA-Z0-9-]*`
var buildIdentifier = `[0-9A-Za-z-]+`
var build = `(?:\+(` + buildIdentifier + `(?:\.` + buildIdentifier + `)*))`
var mainVersion = `(` + numericIdentifier + `)\.(` + numericIdentifier + `)\.(` + numericIdentifier + `)`
var prereleaseIdentifier = `(?:` + numericIdentifier + `|` + nonNumericIdentifier + `)`
var prerelease = `(?:-(` + prereleaseIdentifier + `(?:\.` + prereleaseIdentifier + `)*))`
var fullPlain = `v?` + mainVersion + prerelease + `?` + build + `?`
var reMainVersion = regexp.MustCompile(mainVersion)
var reFull = regexp.MustCompile(`^` + fullPlain + `$`)
var reNumeric = regexp.MustCompile(`^[0-9]+$`)

type Version struct {
	Major      int
	Minor      int
	Patch      int
	Prerelease []string
	Build      []string
	empty      bool
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
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return &Version{empty: true}, nil
	}
	parts := make([]int, 3)
	submatches := reFull.FindStringSubmatch(raw)
	if len(submatches) == 0 {
		return nil, errors.New("invalid version: " + raw)
	}
	for i, s := range submatches[1:4] {
		var err error
		parts[i], err = strconv.Atoi(s)
		if err != nil {
			return nil, err
		}
	}
	return &Version{
		Major:      parts[0],
		Minor:      parts[1],
		Patch:      parts[2],
		Prerelease: removeEmpty(strings.Split(submatches[4], ".")),
		Build:      removeEmpty(strings.Split(submatches[5], ".")),
	}, nil
}

func (this *Version) String() string {
	if this.empty {
		return "*"
	}
	o := fmt.Sprintf("%d.%d.%d", this.Major, this.Minor, this.Patch)
	if len(this.Prerelease) > 0 {
		o = o + "-" + strings.Join(this.Prerelease, ".")
	}
	if len(this.Build) > 0 {
		o = o + "+" + strings.Join(this.Build, ".")
	}
	return o
}

func (this *Version) MarshalJSON() ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(this.String())
	return buffer.Bytes(), err
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
func (a *Version) comparePre(b *Version) int {
	if len(a.Prerelease) > 0 && len(b.Prerelease) == 0 {
		return -1
	} else if len(a.Prerelease) == 0 && len(b.Prerelease) > 0 {
		return 1
	} else if len(a.Prerelease) == 0 && len(b.Prerelease) == 0 {
		return 0
	}

	var i = 0
	for {
		if len(a.Prerelease)-1 < i && len(b.Prerelease)-1 < i {
			return 0
		}
		if len(a.Prerelease)-1 < i {
			return -1
		}
		if len(b.Prerelease)-1 < i {
			return 1
		}
		if a.Prerelease[i] != b.Prerelease[i] {
			return compareIdentifiers(a.Prerelease[i], b.Prerelease[i])
		}
		i++
	}
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
	c = a.comparePatch(b)
	if c != 0 {
		return c
	}
	c = a.comparePre(b)
	if c != 0 {
		return c
	}
	return 0
}

// LT returns true is given version is less than this one
func (a *Version) LT(b *Version) bool {
	return a.compare(b) < 0
}
func (a *Version) LTE(b *Version) bool {
	return a.compare(b) <= 0
}
func (a *Version) GT(b *Version) bool {
	return a.compare(b) > 0
}
func (a *Version) GTE(b *Version) bool {
	return a.compare(b) >= 0
}
func (a *Version) EQ(b *Version) bool {
	return a.compare(b) == 0
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

func compareIdentifiers(a, b string) int {
	anum := reNumeric.MatchString(a)
	bnum := reNumeric.MatchString(b)

	if anum && !bnum {
		return -1
	}
	if !anum && bnum {
		return 1
	}
	if anum && bnum {
		if atoi(a) < atoi(b) {
			return -1
		}
		if atoi(a) > atoi(b) {
			return 1
		}
		return 0
	}
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

func removeEmpty(s []string) []string {
	o := []string{}
	for _, s := range s {
		if s != "" {
			o = append(o, s)
		}
	}
	return o
}
