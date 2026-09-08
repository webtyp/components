package searchbar

import (
	"strings"
	"testing"

	"webtyp.com/widget"
)

func TestSearchBar_RendersBarAndInput(t *testing.T) {
	html := (&SearchBar{}).Render().String()

	for _, want := range []string{"searchbar", "searchbar__icon", "searchbar__glyph", "searchbar__input", "type='search'"} {
		if !strings.Contains(html, want) {
			t.Errorf("markup missing %q\n%s", want, html)
		}
	}
}

func TestSearchBar_NoCustomClearButton(t *testing.T) {
	// type="search" carries its own clear ✕; the bar must not add a second.
	html := (&SearchBar{}).Render().String()
	if strings.Contains(html, "searchbar__clear") || strings.Contains(html, "aria-label='Clear'") {
		t.Errorf("the bar must not render its own clear button — the native type=search ✕ is the one:\n%s", html)
	}
}

func TestSearchBar_DefaultPlaceholder(t *testing.T) {
	html := (&SearchBar{}).Render().String()
	if !strings.Contains(html, "placeholder='Search…'") {
		t.Errorf("expected default placeholder, got\n%s", html)
	}

	html = (&SearchBar{Placeholder: "Buscar..."}).Render().String()
	if !strings.Contains(html, "placeholder='Buscar...'") {
		t.Errorf("expected custom placeholder, got\n%s", html)
	}
	if strings.Contains(html, "Search…") {
		t.Errorf("custom placeholder must not leak the default\n%s", html)
	}
}

func TestSearchBar_IntrinsicWidthCollapsed(t *testing.T) {
	html := (&SearchBar{}).Render().String()
	if !strings.Contains(html, "size='1'") {
		t.Errorf("expected size='1' on the input, got\n%s", html)
	}
}

func TestSearchBar_OnFilterChangeIsOptional(t *testing.T) {
	_ = (&SearchBar{}).Render().String()
}

func TestSearchBar_SatisfiesFilterable(t *testing.T) {
	var f widget.Filterable = &SearchBar{}
	bar := f.(*SearchBar)

	var got string
	f.OnFilterChange(func(term string) { got = term })

	if bar.onFilter == nil {
		t.Fatal("expected OnFilterChange to store the sink")
	}
	bar.onFilter("needle")
	if got != "needle" {
		t.Errorf("stored sink is not the one registered through the interface: got %q", got)
	}
}
