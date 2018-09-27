package semver

import (
	"fmt"
	"regexp"
)

var gtlt = `((?:<|>)?=?)`
var reComparator = regexp.MustCompile(`^` + gtlt + `\s*(` + fullPlain + `)$|^$`)

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

	c := &comparator{}
	if submatches[1] == "=" {
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
	var o string
	if c.eq {
		o = "="
	} else if c.gt {
		o = ">"
	} else if c.gte {
		o = ">="
	} else if c.lte {
		o = "<="
	} else if c.lt {
		o = "<"
	}
	o = o + c.version.String()
	return o
}
