//go:build !wasm

package contentcard

import (
	"regexp"
	"strings"
	"testing"

	. "webtyp.com/dom"
	. "webtyp.com/fmt"
)

var (
	classRegex      = regexp.MustCompile(`\.([a-zA-Z0-9_-]+)`)
	htmlClassRegex  = regexp.MustCompile(`class='([^']*)'`)
	htmlClassRegex2 = regexp.MustCompile(`class="([^"]*)"`)
)

func TestCard_Render(t *testing.T) {
	c := &ContentCard{
		Header: &simpleComponent{html: "Header"},
		Body:   &simpleComponent{html: "Body"},
		Footer: &simpleComponent{html: "Footer"},
	}

	html := c.Render().String()

	if !HasPrefix(html, "<div") {
		t.Error("expected div tag")
	}

	if !Contains(html, "class='contentcard'") {
		t.Error("expected contentcard class")
	}

	if !Contains(html, "class='contentcard__header'") {
		t.Error("expected contentcard__header")
	}
	if !Contains(html, "Header") {
		t.Error("expected Header content")
	}

	if !Contains(html, "class='contentcard__body'") {
		t.Error("expected contentcard__body")
	}
	if !Contains(html, "Body") {
		t.Error("expected Body content")
	}

	if !Contains(html, "class='contentcard__footer'") {
		t.Error("expected contentcard__footer")
	}
	if !Contains(html, "Footer") {
		t.Error("expected Footer content")
	}
}

func TestPairMarkupAndStylesheet(t *testing.T) {
	extractCSSClasses := func(css string) map[string]bool {
		classes := make(map[string]bool)
		matches := classRegex.FindAllStringSubmatch(css, -1)
		for _, m := range matches {
			if len(m) > 1 {
				classes[m[1]] = true
			}
		}
		return classes
	}

	extractHTMLClasses := func(html string) map[string]bool {
		classes := make(map[string]bool)
		matches1 := htmlClassRegex.FindAllStringSubmatch(html, -1)
		for _, m := range matches1 {
			if len(m) > 1 {
				for _, cls := range strings.Fields(m[1]) {
					classes[cls] = true
				}
			}
		}
		matches2 := htmlClassRegex2.FindAllStringSubmatch(html, -1)
		for _, m := range matches2 {
			if len(m) > 1 {
				for _, cls := range strings.Fields(m[1]) {
					classes[cls] = true
				}
			}
		}
		return classes
	}

	filterClasses := func(classes map[string]bool, prefix string) map[string]bool {
		filtered := make(map[string]bool)
		for cls := range classes {
			if strings.HasPrefix(cls, prefix) {
				filtered[cls] = true
			}
		}
		return filtered
	}

	c := &ContentCard{
		Header: &simpleComponent{html: "Header"},
		Body:   &simpleComponent{html: "Body"},
		Footer: &simpleComponent{html: "Footer"},
	}
	html := c.Render().String()
	css := c.RenderCSS().String()

	htmlClasses := filterClasses(extractHTMLClasses(html), "contentcard")
	cssClasses := filterClasses(extractCSSClasses(css), "contentcard")

	for cls := range cssClasses {
		if !htmlClasses[cls] {
			t.Errorf("ContentCard: CSS class %q does not exist in rendered HTML", cls)
		}
	}
	for cls := range htmlClasses {
		if _, ok := cssClasses[cls]; !ok {
			t.Errorf("ContentCard: HTML class %q is unstyled in CSS", cls)
		}
	}
}

type simpleComponent struct {
	html string
}

func (s *simpleComponent) String() string        { return s.html }
func (s *simpleComponent) GetID() string         { return "" }
func (s *simpleComponent) SetID(id string)       {}
func (s *simpleComponent) Children() []Component { return nil }
