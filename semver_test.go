package semver

import (
	"encoding/json"
	"fmt"
	"sort"
	"testing"

	. "github.com/franela/goblin"
)

type testJSON struct {
	Version *Version `json:"version"`
}

func parseJSON(raw string) testJSON {
	var o testJSON
	must(json.Unmarshal([]byte(raw), &o))
	return o
}

func renderJSON(v *Version) string {
	bytes, err := json.Marshal(testJSON{Version: v})
	must(err)
	return string(bytes)
}

func v(s string) *Version {
	return MustParse(s)
}
func r(s string) *Range {
	return MustParseRange(s)
}
func Test(t *testing.T) {
	g := Goblin(t)
	g.Describe("Parse", func() {
		checkRange := func(rawRange, rawVersion string, ok bool) {
			r := MustParseRange(rawRange)
			v := MustParse(rawVersion)
			d := fmt.Sprintf("%s in %s", r, v)
			if !ok {
				d = fmt.Sprintf("%s not in %s", r, v)
			}
			g.It(d, func() {
				g.Assert(r.Valid(v)).Equal(ok)
			})
		}

		checkRange(">1.0.0", "0.1.0", false)
		checkRange(">1.0.0", "1.0.0", false)
		checkRange(">1.0.0", "2.0.0", true)
		checkRange(">=1.0.0", "1.0.0", true)
		checkRange(">=1.0.0", "2.0.0", true)
		checkRange(">=1.0.0", "0.1.0", false)
		checkRange("<1.0.0", "1.0.0", false)
		checkRange("<1.0.0", "0.1.0", true)
		checkRange("<=1.0.0", "0.1.0", true)
		checkRange("<=1.0.0", "1.0.0", true)
		checkRange("<=1.0.0", "1.1.0", false)
		checkRange("=1.0.0", "1.1.0", false)
		checkRange("=1.0.0", "1.0.0", true)
		checkRange("1.0.0", "1.0.0", true)
		checkRange("1.0.0", "1.0.1", false)
		checkRange("1.2.3 - 1.2.4", "1.2.2", false)
		checkRange("1.2.3 - 1.2.4", "1.2.3", true)
		checkRange("1.2.3 - 1.2.4", "1.2.4", true)
		checkRange("1.2.3 - 1.2.4", "1.2.5", false)
		checkRange("> 1.2.3 < 1.2.5", "1.2.2", false)
		checkRange("> 1.2.3 < 1.2.5", "1.2.3", false)
		checkRange("> 1.2.3 < 1.2.5", "1.2.4", true)
		checkRange("> 1.2.3 < 1.2.5", "1.2.5", false)
		checkRange("~ 1.2.3", "1.2.5", true)
		checkRange("~ 1.2.3", "1.3.0", false)
		checkRange("~ 1.2.3", "1.2.2", false)
		checkRange("~ 1.2.3", "1.2.3", true)
		checkRange("^ 1.2.3", "1.2.5", true)
		checkRange("^ 1.2.3", "1.3.0", true)
		checkRange("^ 1.2.3", "2.0.0", false)
		checkRange("^ 1.2.3", "1.2.2", false)
		checkRange("^ 1.2.3", "1.2.3", true)
		checkRange("1.0.0 - 2.0.0", "1.2.3", true)
		checkRange("1.2.3+build", "1.2.3", true)
		checkRange("^1.2.3+build", "1.3.0", true)
		checkRange("1.2.3-pre+asdf - 2.4.3-pre+asdf", "1.2.3", true)
		// checkRange("1.2.3pre+asdf - 2.4.3-pre+asdf", "1.2.3", true, true)
		// checkRange("1.2.3-pre+asdf - 2.4.3pre+asdf", "1.2.3", true, true)
		// checkRange("1.2.3pre+asdf - 2.4.3pre+asdf", "1.2.3", true, true)
		checkRange("1.2.3-pre+asdf - 2.4.3-pre+asdf", "1.2.3-pre.2", true)
		checkRange("1.2.3-pre+asdf - 2.4.3-pre+asdf", "2.4.3-alpha", true)
		checkRange("1.2.3+asdf - 2.4.3+asdf", "1.2.3", true)
		checkRange("1.0.0", "1.0.0", true)
		checkRange(">=*", "0.2.4", true)
		checkRange("", "1.0.0", true)
		checkRange("*", "1.2.3", true)
		// checkRange("*", "v1.2.3", true, true)
		checkRange(">=1.0.0", "1.0.0", true)
		checkRange(">=1.0.0", "1.0.1", true)
		checkRange(">=1.0.0", "1.1.0", true)
		checkRange(">1.0.0", "1.0.1", true)
		checkRange(">1.0.0", "1.1.0", true)
		checkRange("<=2.0.0", "2.0.0", true)
		checkRange("<=2.0.0", "1.9999.9999", true)
		checkRange("<=2.0.0", "0.2.9", true)
		checkRange("<2.0.0", "1.9999.9999", true)
		checkRange("<2.0.0", "0.2.9", true)
		checkRange(">= 1.0.0", "1.0.0", true)
		checkRange(">=  1.0.0", "1.0.1", true)
		checkRange(">=   1.0.0", "1.1.0", true)
		checkRange("> 1.0.0", "1.0.1", true)
		checkRange(">  1.0.0", "1.1.0", true)
		checkRange("<=   2.0.0", "2.0.0", true)
		checkRange("<= 2.0.0", "1.9999.9999", true)
		checkRange("<=  2.0.0", "0.2.9", true)
		checkRange("<    2.0.0", "1.9999.9999", true)
		checkRange("<\t2.0.0", "0.2.9", true)
		// checkRange(">=0.1.97", "v0.1.97", true, true)
		checkRange(">=0.1.97", "0.1.97", true)
		checkRange("0.1.20 || 1.2.4", "1.2.4", true)
		checkRange(">=0.2.3 || <0.0.1", "0.0.0", true)
		checkRange(">=0.2.3 || <0.0.1", "0.2.3", true)
		checkRange(">=0.2.3 || <0.0.1", "0.2.4", true)
		checkRange("||", "1.3.4", true)
		checkRange("2.x.x", "2.1.3", true)
		checkRange("1.2.x", "1.2.3", true)
		checkRange("1.2.x || 2.x", "2.1.3", true)
		checkRange("1.2.x || 2.x", "1.2.3", true)
		checkRange("x", "1.2.3", true)
		checkRange("2.*.*", "2.1.3", true)
		checkRange("1.2.*", "1.2.3", true)
		checkRange("1.2.* || 2.*", "2.1.3", true)
		checkRange("1.2.* || 2.*", "1.2.3", true)
		checkRange("*", "1.2.3", true)
		checkRange("2", "2.1.2", true)
		checkRange("2.3", "2.3.1", true)
		checkRange("~x", "0.0.9", true)
		checkRange("~2", "2.0.9", true)
		checkRange("~2.4", "2.4.0", true)
		checkRange("~2.4", "2.4.5", true)
		checkRange("~>3.2.1", "3.2.2", true)
		checkRange("~1", "1.2.3", true)
		checkRange("~>1", "1.2.3", true)
		checkRange("~> 1", "1.2.3", true)
		checkRange("~1.0", "1.0.2", true)
		checkRange("~ 1.0", "1.0.2", true)
		checkRange("~ 1.0.3", "1.0.12", true)
		checkRange(">=1", "1.0.0", true)
		checkRange(">= 1", "1.0.0", true)
		checkRange("<1.2", "1.1.1", true)
		checkRange("< 1.2", "1.1.1", true)
		checkRange("~v0.5.4-pre", "0.5.5", true)
		checkRange("~v0.5.4-pre", "0.5.4", true)
		checkRange("=0.7.x", "0.7.2", true)
		checkRange("<=0.7.x", "0.7.2", true)
		checkRange(">=0.7.x", "0.7.2", true)
		checkRange("<=0.7.x", "0.6.2", true)
		checkRange("~1.2.1 >=1.2.3", "1.2.3", true)
		checkRange("~1.2.1 =1.2.3", "1.2.3", true)
		checkRange("~1.2.1 1.2.3", "1.2.3", true)
		checkRange("~1.2.1 >=1.2.3 1.2.3", "1.2.3", true)
		checkRange("~1.2.1 1.2.3 >=1.2.3", "1.2.3", true)
		checkRange("~1.2.1 1.2.3", "1.2.3", true)
		checkRange(">=1.2.1 1.2.3", "1.2.3", true)
		checkRange("1.2.3 >=1.2.1", "1.2.3", true)
		checkRange(">=1.2.3 >=1.2.1", "1.2.3", true)
		checkRange(">=1.2.1 >=1.2.3", "1.2.3", true)
		checkRange(">=1.2", "1.2.8", true)
		checkRange("^1.2.3", "1.8.1", true)
		checkRange("^0.1.2", "0.1.2", true)
		checkRange("^0.1", "0.1.2", true)
		checkRange("^0.0.1", "0.0.1", true)
		checkRange("^1.2", "1.4.2", true)
		checkRange("^1.2 ^1", "1.4.2", true)
		checkRange("^1.2.3-alpha", "1.2.3-pre", true)
		checkRange("^1.2.0-alpha", "1.2.0-pre", true)
		checkRange("^0.0.1-alpha", "0.0.1-beta", true)
		checkRange("^0.1.1-alpha", "0.1.1-beta", true)
		checkRange("^x", "1.2.3", true)
		checkRange("x - 1.0.0", "0.9.7", true)
		checkRange("x - 1.x", "0.9.7", true)
		checkRange("1.0.0 - x", "1.9.7", true)
		checkRange("1.x - x", "1.9.7", true)
		checkRange("<=7.x", "7.9.9", true)

		g.It("sorts", func() {
			versions := Versions{v("1.2.3"), v("2.0.0"), v("1.4.2"), v("1.2.4")}
			sort.Sort(versions)
			g.Assert(versions[0]).Equal(v("1.2.3"))
			g.Assert(versions[1]).Equal(v("1.2.4"))
			g.Assert(versions[2]).Equal(v("1.4.2"))
			g.Assert(versions[3]).Equal(v("2.0.0"))

		})
	})
}
