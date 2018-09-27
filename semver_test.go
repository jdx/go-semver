package semver

import (
	"encoding/json"
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

func Test(t *testing.T) {
	g := Goblin(t)
	g.Describe("Parse", func() {
		g.It("parses", func() {
			g.Assert(MustParse("0.0.0")).Equal(&Version{Major: 0, Minor: 0, Patch: 0})
			g.Assert(MustParse("1.2.3")).Equal(&Version{Major: 1, Minor: 2, Patch: 3})
			g.Assert(parseJSON(`{"version": "1.2.3"}`).Version).Equal(&Version{Major: 1, Minor: 2, Patch: 3})
			g.Assert(renderJSON(&Version{Major: 1, Minor: 2, Patch: 3})).Equal(`{"version":"1.2.3"}`)
		})
	})
}
