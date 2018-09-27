package semver

import (
	"fmt"
	"regexp"
	"strings"
)

var gtlt = `((?:<|>)?=?)`
var numericIdentifierLoose = `[0-9]+`
var mainVersionLoose = `(` + numericIdentifierLoose + `)\.` + `(` + numericIdentifierLoose + `)\.` + `(` + numericIdentifierLoose + `)`
var prereleaseIdentifierLoose = `(?:` + numericIdentifierLoose + `|` + nonNumericIdentifier + `)`
var prereleaseLoose = `(?:-?(` + prereleaseIdentifierLoose + `(?:\.` + prereleaseIdentifierLoose + `)*))`
var loosePlain = `[v=\s]*` + mainVersionLoose + prereleaseLoose + `?` + build + `?`
var xRangeIdentifier = numericIdentifier + `|x|X|\*`
var xRangePlain = `[v=\s]*(` + xRangeIdentifier + `)` + `(?:\.(` + xRangeIdentifier + `)` + `(?:\.(` + xRangeIdentifier + `)` + `(?:` + prerelease + `)?` + build + `?` + `)?)?`
var reComparator = regexp.MustCompile(`^` + gtlt + `\s*(` + fullPlain + `)$|^$`)
var reComparatorTrim = regexp.MustCompile(`(\s*)` + gtlt + `\s*(` + loosePlain + `|` + xRangePlain + `)`)
var loneTilde = `(?:~>?)`
var loneCaret = `(?:\^)`
var reTilde = regexp.MustCompile(`^` + loneTilde + xRangePlain + `$`)
var reTildeTrim = regexp.MustCompile(`(\s*)` + loneTilde + `\s+`)
var reCaret = regexp.MustCompile(`^` + loneCaret + xRangePlain + `$`)
var reCaretTrim = regexp.MustCompile(`(\s*)` + loneCaret + `\s+`)
var reHyphenRange = regexp.MustCompile(`^\s*(` + xRangePlain + `)` + `\s+-\s+` + `(` + xRangePlain + `)` + `\s*$`)
var reSpace = regexp.MustCompile(`\s+`)
var reStar = regexp.MustCompile(`(<|>)?=?\s*\*`)
var reXRange = regexp.MustCompile(`^` + gtlt + `\s*` + xRangePlain + `$`)

type comparator struct {
	gt      bool
	gte     bool
	lte     bool
	lt      bool
	eq      bool
	version *Version
}

func parseComparator(raw string) (*comparator, error) {
	var err error
	submatches := reComparator.FindStringSubmatch(raw)
	if len(submatches) == 0 {
		return nil, fmt.Errorf("invalid comparator: %s", raw)
	}
	// fmt.Printf("%q\n", submatches)

	c := &comparator{}
	if submatches[1] == "=" || submatches[1] == "" {
		c.eq = true
	} else if submatches[1] == ">" {
		c.gt = true
	} else if submatches[1] == ">=" {
		c.gte = true
	} else if submatches[1] == "<" {
		c.lt = true
	} else if submatches[1] == "<=" {
		c.lte = true
	}
	c.version, err = Parse(submatches[2])
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (c *comparator) valid(v *Version) bool {
	if c.version.empty {
		return true
	}
	if c.eq {
		return v.EQ(c.version)
	} else if c.gt {
		return v.GT(c.version)
	} else if c.gte {
		return v.GTE(c.version)
	} else if c.lte {
		return v.LTE(c.version)
	} else if c.lt {
		return v.LT(c.version)
	}
	return false
}

func (c *comparator) String() string {
	if c.version.empty {
		return "*"
	}
	var o string
	if c.gt {
		o = ">"
	} else if c.gte {
		o = ">="
	} else if c.lte {
		o = "<="
	} else if c.lt {
		o = "<"
	}
	o = strings.Join([]string{o, c.version.String()}, "")
	return o
}

func replaceCarets(comp string) string {
	comp = strings.TrimSpace(comp)
	elements := []string{}
	for _, comp := range reSpace.Split(comp, -1) {
		elements = append(elements, replaceCaret(comp))
	}
	return strings.Join(elements, " ")
}

func replaceCaret(comp string) string {
	return reCaret.ReplaceAllStringFunc(comp, func(raw string) string {
		submatches := reCaret.FindStringSubmatch(raw)
		M := submatches[1]
		m := submatches[2]
		p := submatches[3]
		pr := submatches[4]
		var ret string

		if isX(M) {
			ret = ""
		} else if isX(m) {
			ret = ">=" + M + ".0.0 <" + itoa(atoi(M)+1) + ".0.0"
		} else if isX(p) {
			if M == "0" {
				ret = ">=" + M + "." + m + ".0 <" + M + "." + itoa(atoi(m)+1) + ".0"
			} else {
				ret = ">=" + M + "." + m + ".0 <" + itoa(atoi(M)+1) + ".0.0"
			}
		} else if pr != "" {
			if pr[0] != '-' {
				pr = "-" + pr
			}
			if M == "0" {
				if m == "0" {
					ret = ">=" + M + "." + m + "." + p + pr +
						" <" + M + "." + m + "." + itoa(atoi(p)+1)
				} else {
					ret = ">=" + M + "." + m + "." + p + pr +
						" <" + M + "." + itoa(atoi(m)+1) + ".0"
				}
			} else {
				ret = ">=" + M + "." + m + "." + p + pr +
					" <" + itoa(atoi(M)+1) + ".0.0"
			}
		} else {
			if M == "0" {
				if m == "0" {
					ret = ">=" + M + "." + m + "." + p +
						" <" + M + "." + m + "." + itoa(atoi(p)+1)
				} else {
					ret = ">=" + M + "." + m + "." + p +
						" <" + M + "." + itoa(atoi(m)+1) + ".0"
				}
			} else {
				ret = ">=" + M + "." + m + "." + p +
					" <" + itoa(atoi(M)+1) + ".0.0"
			}
		}

		return ret
	})
}

func replaceTildes(comp string) string {
	comp = strings.TrimSpace(comp)
	elements := []string{}
	for _, comp := range reSpace.Split(comp, -1) {
		elements = append(elements, replaceTilde(comp))
	}
	return strings.Join(elements, " ")
}

func replaceTilde(comp string) string {
	return reTilde.ReplaceAllStringFunc(comp, func(raw string) string {
		submatches := reTilde.FindStringSubmatch(raw)
		M := submatches[1]
		m := submatches[2]
		p := submatches[3]
		pr := submatches[4]
		var ret string

		if isX(M) {
			ret = ""
		} else if isX(m) {
			ret = ">=" + M + ".0.0 <" + itoa(atoi(M)+1) + ".0.0"
		} else if isX(p) {
			// ~1.2 == >=1.2.0 <1.3.0
			ret = ">=" + M + "." + m + ".0 <" + M + "." + itoa(atoi(m)+1) + ".0"
		} else if pr != "" {
			if pr[0] != '-' {
				pr = "-" + pr
			}
			ret = ">=" + M + "." + m + "." + p + pr +
				" <" + M + "." + itoa(atoi(m)+1) + ".0"
		} else {
			// ~1.2.3 == >=1.2.3 <1.3.0
			ret = ">=" + M + "." + m + "." + p +
				" <" + M + "." + itoa(atoi(m)+1) + ".0"
		}

		return ret
	})
}

func replaceStars(comp string) string {
	comp = strings.TrimSpace(comp)
	return reStar.ReplaceAllString(comp, "")
}

func replaceXRanges(comp string) string {
	comp = strings.TrimSpace(comp)
	elements := []string{}
	for _, comp := range reSpace.Split(comp, -1) {
		elements = append(elements, replaceXRange(comp))
	}
	return strings.Join(elements, " ")
}

func replaceXRange(comp string) string {
	comp = strings.TrimSpace(comp)
	return reXRange.ReplaceAllStringFunc(comp, func(raw string) string {
		ret := raw
		submatches := reXRange.FindStringSubmatch(raw)
		gtlt := submatches[1]
		M := submatches[2]
		m := submatches[3]
		p := submatches[4]
		// pr := submatches[4]
		xM := isX(M)
		xm := xM || isX(m)
		xp := xm || isX(p)
		anyX := xp

		if gtlt == "=" && anyX {
			gtlt = ""
		}

		if xM {
			if gtlt == ">" || gtlt == "<" {
				// nothing is allowed
				ret = "<0.0.0"
			} else {
				// nothing is forbidden
				ret = "*"
			}
		} else if gtlt != "" && anyX {
			// replace X with 0
			if xm {
				m = "0"
			}
			if xp {
				p = "0"
			}

			if gtlt == ">" {
				// >1 => >=2.0.0
				// >1.2 => >=1.3.0
				// >1.2.3 => >= 1.2.4
				gtlt = ">="
				if xm {
					M = itoa(atoi(M) + 1)
					m = "0"
					p = "0"
				} else if xp {
					m = itoa(atoi(m) + 1)
					p = "0"
				}
			} else if gtlt == "<=" {
				// <=0.7.x is actually <0.8.0, since any 0.7.x should
				// pass.  Similarly, <=7.x is actually <8.0.0, etc.
				gtlt = "<"
				if xm {
					M = itoa(atoi(M) + 10)
				} else {
					m = itoa(atoi(m) + 1)
				}
			}

			ret = gtlt + M + "." + m + "." + p
		} else if xm {
			ret = ">=" + M + ".0.0 <" + itoa(atoi(M)+1) + ".0.0"
		} else if xp {
			ret = ">=" + M + "." + m + ".0 <" + M + "." + itoa(atoi(m)+1) + ".0"
		}

		return ret
	})
}
