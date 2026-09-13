//go:build !wasm

package decktabs

import (
	"regexp"
	"strings"
	"testing"

	. "webtyp.com/dom"
	"webtyp.com/widget"
)

// panelStub is a minimal dom.Component standing in for real panel content.
type panelStub struct{ html string }

func (p *panelStub) String() string        { return p.html }
func (p *panelStub) GetID() string         { return "" }
func (p *panelStub) SetID(id string)       {}
func (p *panelStub) Children() []Component { return nil }

func threeItems() []Item {
	return []Item{
		{ID: "data", Label: "Datos", Panel: &panelStub{html: "PANEL-DATA"}},
		{ID: "hours", Label: "Horario", Panel: &panelStub{html: "PANEL-HOURS"}},
		{ID: "services", Label: "Servicios", Panel: &panelStub{html: "PANEL-SERVICES"}},
	}
}

// currentAttr is the rendered form of the widget.Current state attribute. It is
// derived from the State itself, never written by hand — the state owns its own
// encoding.
func currentAttr() string {
	return widget.Current.Key() + "='" + widget.Current.Value() + "'"
}

func TestDeckTabs_RendersOneTabAndPanelPerItem(t *testing.T) {
	tb := &DeckTabs{Items: threeItems()}
	html := tb.Render().String()

	if got := strings.Count(html, "role='tab'"); got != 3 {
		t.Errorf("got %d tabs, want 3", got)
	}
	if got := strings.Count(html, "role='tabpanel'"); got != 3 {
		t.Errorf("got %d panels, want 3", got)
	}
	if !strings.Contains(html, "role='tablist'") {
		t.Error("expected a tablist")
	}
}

func TestDeckTabs_FirstItemActiveByDefault(t *testing.T) {
	tb := &DeckTabs{Items: threeItems()}
	html := tb.Render().String()

	if tb.Active.Get() != "data" {
		t.Fatalf("Active = %q, want %q", tb.Active.Get(), "data")
	}
	// Exactly two nodes carry Current: the first tab and the first panel.
	if got := strings.Count(html, currentAttr()); got != 2 {
		t.Errorf("got %d nodes carrying Current, want 2 (tab + panel)", got)
	}
}

func TestDeckTabs_RespectsInjectedActive(t *testing.T) {
	active := NewString("hours")
	tb := &DeckTabs{Items: threeItems(), Active: active}
	html := tb.Render().String()

	if active.Get() != "hours" {
		t.Fatalf("injected Active was overwritten: %q", active.Get())
	}
	if got := strings.Count(html, currentAttr()); got != 2 {
		t.Errorf("got %d nodes carrying Current, want 2", got)
	}

	// Current must sit on the "hours" pair, not on the first item. Checked by
	// data-id, never by a composed element id — dom mints those.
	for _, role := range []string{"tab", "tabpanel"} {
		for _, el := range attrsOf(html, role) {
			isCurrent := el[widget.Current.Key()] == widget.Current.Value()
			if want := el["data-id"] == "hours"; isCurrent != want {
				t.Errorf("%s %q: Current=%v, want %v", role, el["data-id"], isCurrent, want)
			}
		}
	}
}

var attrRe = regexp.MustCompile(`([a-z-]+)='([^']*)'`)

// attrsOf returns the attribute map of every element in html whose attributes
// include role=want. Parsing the markup, rather than asserting on composed id
// strings, is the point: the ids are MINTED by dom, so a test that spells one
// out would be asserting the very thing this component must not do.
func attrsOf(html, want string) []map[string]string {
	var out []map[string]string
	for _, chunk := range strings.Split(html, "<")[1:] {
		end := strings.Index(chunk, ">")
		if end < 0 {
			continue
		}
		attrs := map[string]string{}
		for _, m := range attrRe.FindAllStringSubmatch(chunk[:end], -1) {
			attrs[m[1]] = m[2]
		}
		if attrs["role"] == want {
			out = append(out, attrs)
		}
	}
	return out
}

func TestDeckTabs_AriaLinksTabToPanel(t *testing.T) {
	tb := &DeckTabs{Items: threeItems()}
	html := tb.Render().String()

	tabs := attrsOf(html, "tab")
	panels := attrsOf(html, "tabpanel")
	if len(tabs) != 3 || len(panels) != 3 {
		t.Fatalf("got %d tabs and %d panels, want 3 and 3", len(tabs), len(panels))
	}

	panelByID := map[string]map[string]string{}
	for _, p := range panels {
		if p["id"] == "" {
			t.Fatal("a panel rendered with no id — GetID must mint one")
		}
		panelByID[p["id"]] = p
	}

	for i, tab := range tabs {
		if tab["id"] == "" {
			t.Fatalf("tab %d rendered with no id — GetID must mint one", i)
		}
		panel, ok := panelByID[tab["aria-controls"]]
		if !ok {
			t.Errorf("tab %q points at aria-controls=%q, which is no panel on this page",
				tab["data-id"], tab["aria-controls"])
			continue
		}
		if panel["aria-labelledby"] != tab["id"] {
			t.Errorf("panel %q is labelled by %q, want its own tab %q",
				panel["data-id"], panel["aria-labelledby"], tab["id"])
		}
		if panel["data-id"] != tab["data-id"] {
			t.Errorf("tab %q is wired to panel %q", tab["data-id"], panel["data-id"])
		}
	}
}

// Two DeckTabs on one page, deliberately sharing every Item.ID. Composing an
// id from Item.ID made this render emit each id twice, which dom's claimID
// panics on — the crash this test exists to keep away.
func TestDeckTabs_TwoSetsSharingItemIDsDoNotCollide(t *testing.T) {
	first := (&DeckTabs{Items: threeItems()}).Render().String()
	second := (&DeckTabs{Items: threeItems()}).Render().String()

	seen := map[string]bool{}
	for _, role := range []string{"tab", "tabpanel"} {
		for _, el := range append(attrsOf(first, role), attrsOf(second, role)...) {
			id := el["id"]
			if id == "" {
				t.Fatalf("%s rendered with no id", role)
			}
			if seen[id] {
				t.Errorf("id %q emitted by two nodes — dom resolves handlers and "+
					"signal patches by id, so the second one would be dead", id)
			}
			seen[id] = true
		}
	}
}

func TestDeckTabs_TabsAreTypeButton(t *testing.T) {
	tb := &DeckTabs{Items: threeItems()}
	html := tb.Render().String()

	// A bare <button> defaults to type=submit; a tab set inside a form would
	// then submit it on every tab change.
	if got := strings.Count(html, "type='button'"); got != 3 {
		t.Errorf("got %d type=button, want 3", got)
	}
}

func TestDeckTabs_AllPanelsStayMounted(t *testing.T) {
	tb := &DeckTabs{Items: threeItems()}
	html := tb.Render().String()

	// SlideDeck's contract: the state decides what is on screen, nothing is
	// unmounted. Inactive panel content must be present in the markup.
	for _, want := range []string{"PANEL-DATA", "PANEL-HOURS", "PANEL-SERVICES"} {
		if !strings.Contains(html, want) {
			t.Errorf("panel content %q is missing — panels must stay mounted", want)
		}
	}
}

func TestDeckTabs_AriaLabelOnlyWhenSet(t *testing.T) {
	plain := (&DeckTabs{Items: threeItems()}).Render().String()
	if strings.Contains(plain, "aria-label=") {
		t.Error("no aria-label expected when Label is empty")
	}
	named := (&DeckTabs{Items: threeItems(), Label: "Personal"}).Render().String()
	if !strings.Contains(named, "aria-label='Personal'") {
		t.Error("expected the Label to become the tablist aria-label")
	}
}

func TestDeckTabs_ActivateIsANoOpOnTheActiveTab(t *testing.T) {
	calls := 0
	tb := &DeckTabs{Items: threeItems(), OnChange: func(string) { calls++ }}
	tb.Init(nil)

	tb.activate("data") // already active
	if calls != 0 {
		t.Errorf("OnChange fired %d times re-selecting the active tab, want 0", calls)
	}
	tb.activate("hours")
	if calls != 1 {
		t.Errorf("OnChange fired %d times, want 1", calls)
	}
	if tb.Active.Get() != "hours" {
		t.Errorf("Active = %q, want %q", tb.Active.Get(), "hours")
	}
}

func TestDeckTabs_EmptyItemsRendersNothingAndDoesNotPanic(t *testing.T) {
	tb := &DeckTabs{}
	html := tb.Render().String()

	if strings.Contains(html, "role='tab'") || strings.Contains(html, "role='tablist'") {
		t.Error("no items must render no strip and no tabs")
	}
}

func TestDeckTabs_RenderCSSNotNil(t *testing.T) {
	// RenderCSS panics on an invalid composition, so this also asserts the
	// stylesheet validates.
	if sheet := (&DeckTabs{}).RenderCSS(); sheet == nil {
		t.Error("RenderCSS returned nil")
	}
}

func TestDeckTabs_WidgetKindIsTabs(t *testing.T) {
	if k := (&DeckTabs{}).WidgetKind(); k != widget.Tabs {
		t.Errorf("WidgetKind = %v, want widget.Tabs", k)
	}
}
