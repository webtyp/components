package selectsearch

import (
	"regexp"
	"testing"

	"webtyp.com/dom"
)

var _ dom.Component = (*SelectSearch)(nil)

var idAttr = regexp.MustCompile(`(id|for|aria-controls)='[^']*'`)

func TestSelectSearch_InitTwiceSafe(t *testing.T) {
	c := &SelectSearch{Options: []SsOption{{ID: "1", Label: "One"}}}

	c.Init(nil)
	c.Init(nil)
}

// Render is idempotent in shape: each call may assign fresh auto-ids
// (e.g. via BindChildren), so ids are normalized before comparing.
func TestSelectSearch_RenderIdempotent(t *testing.T) {
	c := &SelectSearch{Options: []SsOption{{ID: "1", Label: "One"}}}
	c.Init(nil)

	first := idAttr.ReplaceAllString(c.Render().String(), "id='_'")
	second := idAttr.ReplaceAllString(c.Render().String(), "id='_'")

	if first != second {
		t.Errorf("Render() not idempotent in shape: first=%q second=%q", first, second)
	}
}
