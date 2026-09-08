//go:build !wasm

package selectsearch

import (
	"strings"
	"testing"
)

func TestSheetValidates(t *testing.T) {
	c := &SelectSearch{}
	c.Init(nil)
	if errs := c.sheet().Validate(); len(errs) > 0 {
		t.Errorf("selectsearch sheet must validate, got:\n%v", errs)
	}
}

func TestHeaderKeepsItsShapeWhenNarrow(t *testing.T) {
	// Row() carries flex-wrap: wrap. Without KeepSize on both the bar and its
	// cap a narrow viewport wrapped the text under the square. searchbar
	// already answers this; the header must answer it identically.
	s := (&SelectSearch{}).RenderCSS().String()
	for _, want := range []string{
		".selectsearch__header",
		"flex-shrink: 0;",
		"min-height: var(--control-height",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("expected %q in the header/icon rules, got:\n%s", want, s)
		}
	}
}

func TestChevronTurnsOnOpen(t *testing.T) {
	s := (&SelectSearch{}).RenderCSS().String()
	if !strings.Contains(s, "transform: rotate(0deg);") {
		t.Errorf("expected the chevron's resting rotation, got:\n%s", s)
	}
	if !strings.Contains(s, `.selectsearch[data-open="true"] .selectsearch__glyph`) {
		t.Errorf("expected the open state to reach the glyph, got:\n%s", s)
	}
	if !strings.Contains(s, "transform: rotate(180deg);") {
		t.Errorf("expected the open rotation, got:\n%s", s)
	}
}

func TestDropdownPadsItsClippedCorners(t *testing.T) {
	// As(Panel) brings RadiusMd and HideOverflow clips to it: a child flush
	// against the edge loses its corners AND its inset focus ring.
	s := (&SelectSearch{}).RenderCSS().String()
	i := strings.Index(s, ".selectsearch__dropdown {")
	if i < 0 {
		t.Fatal("expected a dropdown rule")
	}
	b := s[i:]
	if e := strings.Index(b, "}"); e > 0 {
		b = b[:e]
	}
	if !strings.Contains(b, "padding:") {
		t.Errorf("the clipping dropdown must pad its children off the rounded corners, got:\n%s", b)
	}
}

func TestOptionsWearTheChassisHoverAndSelection(t *testing.T) {
	// The same amber language targetlist uses. A grey hover made the dropdown
	// read as a foreign piece bolted onto the app.
	s := (&SelectSearch{}).RenderCSS().String()
	for _, want := range []string{
		"--color-accent-wash", // hover + focus
		`.selectsearch__option[data-selected="true"]`,
		"--color-accent", // selected + press
	} {
		if !strings.Contains(s, want) {
			t.Errorf("expected %q, got:\n%s", want, s)
		}
	}
}

func TestOptionsReuseTheSharedListContainer(t *testing.T) {
	// listgap owns the list rhythm for the whole component set. Re-declaring
	// it here is what let the dropdown drift into a third spacing.
	s := (&SelectSearch{}).RenderCSS().String()
	if !strings.Contains(s, ".selectsearch__options") || !strings.Contains(s, "overflow-y: auto;") {
		t.Errorf("expected the shared list container on the options, got:\n%s", s)
	}
}
