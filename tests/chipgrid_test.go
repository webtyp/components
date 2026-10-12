//go:build !wasm

package components_test

import (
	"strings"
	"testing"

	"webtyp.com/components/chipgrid"
	"webtyp.com/dom"
	"webtyp.com/lang"
	"webtyp.com/widget"
)

func hours() []chipgrid.Chip {
	return []chipgrid.Chip{
		{Value: "09:00", Label: "09:00"},
		{Value: "09:30", Label: "09:30", Disabled: true},
		{Value: "10:00", Label: "10:00"},
	}
}

func newGrid(chips []chipgrid.Chip) *chipgrid.ChipGrid {
	g := &chipgrid.ChipGrid{}
	g.Init(nil)
	g.SetChips(chips)
	return g
}

// chipMarkup returns the rendered <button> of each chip, in order.
func chipMarkup(g *chipgrid.ChipGrid) []string {
	parts := strings.Split(g.Render().String(), "<button")
	return parts[1:]
}

func TestChipGrid_EmptyShowsDefaultText(t *testing.T) {
	out := newGrid(nil).Render().String()
	if !strings.Contains(out, "class='chipgrid__empty'") || !strings.Contains(out, ">No options available<") {
		t.Errorf("expected the default empty text, got %s", out)
	}
	if strings.Contains(out, "<button") {
		t.Errorf("no chips expected, got %s", out)
	}
}

func TestChipGrid_EmptyShowsCustomText(t *testing.T) {
	if err := lang.Load([]byte(`{"languages":["es"],"keys":{"No times available":["No hay horas disponibles"]}}`)); err != nil {
		t.Fatal(err)
	}
	lang.OutLang(lang.ES)
	defer lang.OutLang(lang.EN)
	g := &chipgrid.ChipGrid{Empty: "No times available"}
	g.Init(nil)
	g.SetChips(nil)
	if out := g.Render().String(); !strings.Contains(out, ">No hay horas disponibles<") {
		t.Errorf("expected the translated custom empty text, got %s", out)
	}
}

func TestChipGrid_RendersOneOptionPerChip(t *testing.T) {
	g := newGrid(hours())
	out := g.Render().String()
	if !strings.Contains(out, "role='listbox'") {
		t.Error("root must be a listbox")
	}
	if got := strings.Count(out, "role='option'"); got != 3 {
		t.Fatalf("got %d options, want 3", got)
	}
	chips := chipMarkup(g)
	for i, want := range []string{"09:00", "09:30", "10:00"} {
		if !strings.Contains(chips[i], ">"+want+"<") {
			t.Errorf("chip %d: missing label %s", i, want)
		}
	}
	if !strings.Contains(chips[1], stateAttr(widget.Disabled)) || !strings.Contains(chips[1], "disabled") {
		t.Error("a disabled chip carries the Disabled state and the disabled attribute")
	}
	if strings.Contains(chips[0], stateAttr(widget.Disabled)) {
		t.Error("an enabled chip must not carry Disabled")
	}
	if strings.Contains(out, "chipgrid__empty") {
		t.Error("the empty text must not show when there are chips")
	}
}

func TestChipGrid_SetChipsKeepsEnabledSelection(t *testing.T) {
	g := newGrid(hours())
	g.Selected.Set("10:00")
	g.SetChips(hours())
	if g.Selected.Get() != "10:00" {
		t.Errorf("Selected = %q, want 10:00", g.Selected.Get())
	}
	chips := chipMarkup(g)
	if !strings.Contains(chips[2], stateAttr(widget.Selected)) || !strings.Contains(chips[2], "aria-selected='true'") {
		t.Error("the selected chip carries Selected and aria-selected='true'")
	}
	if !strings.Contains(chips[0], "aria-selected='false'") {
		t.Error("an unselected chip carries aria-selected='false'")
	}
}

func TestChipGrid_SetChipsClearsMissingOrDisabledSelection(t *testing.T) {
	g := newGrid(hours())
	g.Selected.Set("11:00")
	g.SetChips(hours())
	if g.Selected.Get() != "" {
		t.Errorf("missing value: Selected = %q, want empty", g.Selected.Get())
	}
	g.Selected.Set("09:30")
	g.SetChips(hours())
	if g.Selected.Get() != "" {
		t.Errorf("disabled value: Selected = %q, want empty", g.Selected.Get())
	}
}

func TestChipGrid_TabStopOnSelectedOrFirstEnabled(t *testing.T) {
	chips := []chipgrid.Chip{
		{Value: "a", Label: "a", Disabled: true},
		{Value: "b", Label: "b"},
		{Value: "c", Label: "c"},
	}
	g := newGrid(chips)
	m := chipMarkup(g)
	if !strings.Contains(m[1], "tabindex='0'") || !strings.Contains(m[2], "tabindex='-1'") || !strings.Contains(m[0], "tabindex='-1'") {
		t.Errorf("none selected: the first ENABLED chip is the tab stop, got %v", m)
	}
	g.Selected.Set("c")
	g.SetChips(chips)
	m = chipMarkup(g)
	if !strings.Contains(m[2], "tabindex='0'") || !strings.Contains(m[1], "tabindex='-1'") {
		t.Errorf("selected c: c is the tab stop, got %v", m)
	}
}

func TestChipGrid_UsesHostSelectedSignal(t *testing.T) {
	sel := dom.NewString("09:00")
	g := &chipgrid.ChipGrid{Selected: sel}
	g.Init(nil)
	g.SetChips(hours())
	if g.Selected != sel || sel.Get() != "09:00" {
		t.Error("ChipGrid must keep the host's Selected signal")
	}
}

func TestChipGrid_RenderWithoutInit(t *testing.T) {
	g := &chipgrid.ChipGrid{}
	if out := g.Render().String(); !strings.Contains(out, "chipgrid__empty") {
		t.Errorf("Render before Init/SetChips shows the empty text, got %s", out)
	}
	g.SetChips(hours()) // SetChips before Init is allowed too
	if g.Selected == nil {
		t.Fatal("Selected must be created")
	}
}

func TestChipGrid_RenderCSS(t *testing.T) {
	sheet := (&chipgrid.ChipGrid{}).RenderCSS()
	if sheet == nil || sheet.String() == "" {
		t.Fatal("empty stylesheet")
	}
}
