package searchbar

import (
	"strings"
	"testing"

	"webtyp.com/fmt/lang"
	"webtyp.com/widget"
)

func TestSearchBar_RendersBarAndInput(t *testing.T) {
	html := (&SearchBar{}).Render().String()

	// No searchbar__glyph: the magnifier <svg> carries no class of its own,
	// because PartIcon's IconCap() sizes it as a child rule. A glyph class
	// here would only exist to re-answer a question the cap already owns.
	for _, want := range []string{"searchbar", "searchbar__icon", "searchbar__input", "type='search'", "<svg"} {
		if !strings.Contains(html, want) {
			t.Errorf("markup missing %q\n%s", want, html)
		}
	}
	if strings.Contains(html, "searchbar__glyph") {
		t.Errorf("the glyph must not carry a class — IconCap sizes it:\n%s", html)
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

// TestSearchBar_DefaultPlaceholderIsTranslatable: the default is THIS library's
// own chrome text, so it goes through lang like every other label here — a bar
// that reads "Search…" inside a Spanish app is the one English word left on the
// screen. The host's own Placeholder is parameterized input and stays verbatim:
// translating what the app passed in would be this library deciding the app's
// language. Registering a dictionary from a test is how a consumer is simulated
// (see layout/AGENTS.md — a library never calls RegisterWords in production).
func TestSearchBar_DefaultPlaceholderIsTranslatable(t *testing.T) {
	lang.RegisterWords([]lang.DictEntry{{EN: "Search…", ES: "Buscar…"}})
	lang.OutLang(lang.ES)
	defer lang.OutLang(lang.EN)

	html := (&SearchBar{}).Render().String()
	if !strings.Contains(html, "placeholder='Buscar…'") {
		t.Errorf("the default placeholder must render through the consumer's dictionary, got\n%s", html)
	}

	html = (&SearchBar{Placeholder: "Filtrar por RUT"}).Render().String()
	if !strings.Contains(html, "placeholder='Filtrar por RUT'") {
		t.Errorf("a host-supplied placeholder is parameterized input and must render verbatim, got\n%s", html)
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
