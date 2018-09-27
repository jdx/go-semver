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
		checkRange := func(r, v string, ok bool) {
			d := fmt.Sprintf("%s in %s", r, v)
			if !ok {
				d = fmt.Sprintf("%s not in %s", r, v)
			}
			g.It(d, func() {
				g.Assert(MustParseRange(r).Valid(MustParse(v))).Equal(ok)
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
