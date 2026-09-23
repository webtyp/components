//go:build !wasm

package selectsearch

import (
	"regexp"
	"strings"
	"testing"

	"webtyp.com/fmt"
	"webtyp.com/fmt/lang"
	"webtyp.com/widget"
)

var (
	classRegex      = regexp.MustCompile(`\.([a-zA-Z0-9_-]+)`)
	htmlClassRegex  = regexp.MustCompile(`class='([^']*)'`)
	htmlClassRegex2 = regexp.MustCompile(`class="([^"]*)"`)
)

func TestSelectSearch_Render(t *testing.T) {
	c := &SelectSearch{
		Placeholder: "Choose category",
		Options: []SsOption{
			{ID: "1", Label: "Automobiles", Description: "auto"},
			{ID: "2", Label: "Film & Animation", Description: "anime"},
		},
	}
	c.Init(nil)

	html := c.Render().String()

	if !fmt.Contains(html, "selectsearch") {
		t.Error("expected selectsearch class")
	}
	if !fmt.Contains(html, "selectsearch__toggle") {
		t.Error("expected selectsearch__toggle checkbox")
	}
	if !fmt.Contains(html, "Choose category") {
		t.Error("expected placeholder text")
	}
	// With build-once Show, the dropdown is always serialized; the container carries
	// display:none when isOpen is false so options are present in the SSR HTML.
}

func manyOptions(n int) []SsOption {
	opts := make([]SsOption, n)
	for i := range opts {
		opts[i] = SsOption{ID: fmt.Sprintf("id-%d", i), Label: fmt.Sprintf("Option %d", i)}
	}
	return opts
}

func TestSearchAutoHidesTheFieldOnAShortList(t *testing.T) {
	// The field is what summons the phone's on-screen keyboard the moment the
	// picker opens. On a list of five names there is nothing to type into and
	// the keyboard covers the very rows the user came to read.
	c := &SelectSearch{Options: manyOptions(5)}
	c.Init(nil)

	if c.searchVisible() {
		t.Errorf("SearchAuto must hide the field for %d options (threshold %d)", 5, searchThreshold)
	}
	if c.searchShown.Get() {
		t.Error("searchShown signal must agree with searchVisible at Init")
	}
}

func TestSearchAutoShowsTheFieldOnALongList(t *testing.T) {
	c := &SelectSearch{Options: manyOptions(searchThreshold)}
	c.Init(nil)

	if !c.searchVisible() {
		t.Errorf("SearchAuto must show the field at %d options", searchThreshold)
	}
}

func TestSearchAutoShowsTheFieldWhenARemoteSourceIsWired(t *testing.T) {
	// A consumer that fetches results for a term has already declared the list
	// is not browsable, whatever len(Options) says at this instant.
	c := &SelectSearch{
		Options:  manyOptions(2),
		OnSearch: func(string) []SsOption { return nil },
	}
	c.Init(nil)

	if !c.searchVisible() {
		t.Error("SearchAuto must show the field when OnSearch is wired, however short the list")
	}
}

func TestSearchModeOverridesTheCount(t *testing.T) {
	always := &SelectSearch{Options: manyOptions(1), Search: SearchAlways}
	always.Init(nil)
	if !always.searchVisible() {
		t.Error("SearchAlways must show the field on a one-item list")
	}

	never := &SelectSearch{Options: manyOptions(500), Search: SearchNever}
	never.Init(nil)
	if never.searchVisible() {
		t.Error("SearchNever must hide the field however long the list")
	}
}

func TestSetOptionsMovesTheSearchDecision(t *testing.T) {
	// A picker rendered empty and filled from an async source must be able to
	// grow its field: the count is what SearchAuto decides on.
	c := &SelectSearch{}
	c.Init(nil)
	if c.searchShown.Get() {
		t.Error("an empty picker must start without a field")
	}

	c.SetOptions(manyOptions(searchThreshold + 5))

	if !c.searchShown.Get() {
		t.Error("SetOptions must re-decide searchShown, or a late-arriving list never gains its field")
	}
}

func TestSelectSearch_SelectedValue(t *testing.T) {
	c := &SelectSearch{
		Placeholder: "Choose category",
	}
	c.Init(nil)
	c.selectedLabel.Set("Automobiles")

	html := c.Render().String()
	if !fmt.Contains(html, "Automobiles") {
		t.Error("expected selected label")
	}
}

func TestHeaderEchoesThePickedRow(t *testing.T) {
	// The collapsed header must carry the whole picked option — name, second
	// line and trailing datum — using the same PartText/PartLabel/PartSublabel
	// /PartDesc an open row uses, so the two never drift.
	c := &SelectSearch{
		Placeholder: "Seleccione un paciente...",
		Options: []SsOption{
			{ID: "p3", Label: "Diego Rojas, 58 años", Sublabel: "9.876.543-2", Description: "10:15"},
		},
	}
	c.Init(nil)

	// Before any choice: the placeholder line, nothing else.
	empty := c.Render().String()
	if !fmt.Contains(empty, "Seleccione un paciente...") {
		t.Error("empty header must show the placeholder")
	}
	if !fmt.Contains(empty, string(ClsSsPlaceholder)) {
		t.Errorf("empty header must render the %s part", ClsSsPlaceholder)
	}

	c.selectOption(c.Options[0])
	picked := c.Render().String()

	for _, want := range []string{"Diego Rojas, 58 años", "9.876.543-2", "10:15"} {
		if !fmt.Contains(picked, want) {
			t.Errorf("picked header missing %q", want)
		}
	}
	// The header body must use the shared option parts, not header-only ones.
	for _, cls := range []string{string(ClsSsHeaderBody), string(ClsSsText), string(ClsSsLabel), string(ClsSsSublabel), string(ClsSsDesc)} {
		if !fmt.Contains(picked, cls) {
			t.Errorf("picked header missing shared part class %q", cls)
		}
	}
}

func TestSelectSearch_OpenState_RendersChecked(t *testing.T) {
	c := &SelectSearch{
		Options: []SsOption{
			{ID: "1", Label: "Apple"},
		},
	}
	c.Init(nil)
	c.isOpen.Set(true)

	el := c.Render()
	html := el.String()
	if !fmt.Contains(html, "checked") {
		t.Error("expected 'checked' attribute on toggle when isOpen=true")
	}
}

func TestSelectSearch_Filtering(t *testing.T) {
	c := &SelectSearch{
		Options: []SsOption{
			{ID: "1", Label: "Automobiles"},
			{ID: "2", Label: "Film & Animation"},
		},
	}
	c.Init(nil)
	c.isOpen.Set(true)
	c.query.Set("Film")
	// query.Set doesn't trigger OnChange in standard tests unless manually called or using gotest/WASM
	c.rows.Set(c.buildRows("Film"))

	el := c.Render()
	html := el.String()
	if fmt.Contains(html, "Automobiles") {
		t.Error("expected Automobiles to be filtered out")
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

	ss := &SelectSearch{}
	ss.Init(nil)
	ss.isOpen.Set(true)
	ss.SetOptions([]SsOption{{ID: "1", Label: "A", Sublabel: "C", Description: "B"}})
	html := ss.Render().String()
	// Render option row too
	html += ss.buildRows("")[0].String()

	css := ss.RenderCSS().String()

	htmlClasses := filterClasses(extractHTMLClasses(html), "selectsearch")
	cssClasses := filterClasses(extractCSSClasses(css), "selectsearch")

	for cls := range cssClasses {
		if !htmlClasses[cls] {
			t.Errorf("SelectSearch CSS class %q does not exist in rendered HTML", cls)
		}
	}
	for cls := range htmlClasses {
		if !cssClasses[cls] {
			t.Errorf("SelectSearch HTML class %q is unstyled in CSS", cls)
		}
	}
}

func TestTwoPickersDoNotShareIDs(t *testing.T) {
	a, b := &SelectSearch{}, &SelectSearch{}
	a.Init(nil)
	b.Init(nil)
	ha, hb := a.Render().String(), b.Render().String()
	if strings.Contains(ha, "ss-toggle") || strings.Contains(hb, "ss-toggle") {
		t.Error("ids must be per-instance, not the fixed ss-* literals")
	}
	if ha == hb {
		t.Error("two pickers rendered identical markup — their ids collide")
	}
}

func TestSelectSearch_SatisfiesFilterable(t *testing.T) {
	var _ widget.Filterable = (*SelectSearch)(nil)
}

func TestSelectSearch_OnFilterChange_FiresOnSelection(t *testing.T) {
	c := &SelectSearch{Options: []SsOption{
		{ID: "p1", Label: "Juan Pérez", Description: "09:00"},
	}}
	c.Init(nil)

	var got string
	fired := 0
	c.OnFilterChange(func(term string) {
		got = term
		fired++
	})

	c.selectOption(c.Options[0])

	if fired != 1 {
		t.Fatalf("expected OnFilterChange to fire exactly once, got %d", fired)
	}
	if got != "p1" {
		t.Fatalf("expected OnFilterChange term %q, got %q", "p1", got)
	}
}

func TestSelectSearch_OnSelect_StillFiresAlongsideFilterable(t *testing.T) {
	c := &SelectSearch{Options: []SsOption{
		{ID: "p1", Label: "Juan Pérez", Description: "09:00"},
	}}
	c.Init(nil)

	var gotID, gotDesc string
	c.OnSelect = func(id, description string) {
		gotID, gotDesc = id, description
	}
	c.OnFilterChange(func(string) {}) // both sinks registered — neither must suppress the other

	c.selectOption(c.Options[0])

	if gotID != "p1" || gotDesc != "09:00" {
		t.Fatalf("expected OnSelect(%q, %q), got (%q, %q)", "p1", "09:00", gotID, gotDesc)
	}
}

// TestSelectSearch_SearchPlaceholderIsTranslatable: the open panel's search
// field is this library's own chrome and must read in the consumer's language —
// a picker that says "Search…" inside a Spanish app is the one English word on
// screen. It deliberately shares components/searchbar's key so an app registers
// the translation once, not once per component.
func TestSelectSearch_SearchPlaceholderIsTranslatable(t *testing.T) {
	lang.RegisterWords([]lang.DictEntry{{EN: "Search…", ES: "Buscar…"}})
	lang.OutLang(lang.ES)
	defer lang.OutLang(lang.EN)

	html := (&SelectSearch{Placeholder: "Seleccione un paciente"}).Render().String()
	if !strings.Contains(html, "placeholder='Buscar…'") {
		t.Errorf("the panel's search placeholder must render through the dictionary, got\n%s", html)
	}
	if !strings.Contains(html, "Seleccione un paciente") {
		t.Errorf("the host-supplied Placeholder is parameterized input and must survive verbatim, got\n%s", html)
	}
}
