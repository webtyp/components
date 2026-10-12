//go:build !wasm

package components_test

import (
	"strings"
	"testing"

	"webtyp.com/components/chipgrid"
	"webtyp.com/dom"
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

// Chip markup lives behind BindChildren, which only the browser renders: every
// assertion on it is in chipgrid_ui_wasm_test.go. Here: selection logic and CSS.

func TestChipGrid_SetChipsKeepsEnabledSelection(t *testing.T) {
	g := newGrid(hours())
	g.Selected.Set("10:00")
	g.SetChips(hours())
	if g.Selected.Get() != "10:00" {
		t.Errorf("Selected = %q, want 10:00", g.Selected.Get())
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
	if out := g.Render().String(); !strings.Contains(out, "role='listbox'") {
		t.Errorf("Render before Init/SetChips renders the listbox root, got %s", out)
	}
	g.SetChips(hours()) // SetChips before Init is allowed too
	if g.Selected == nil {
		t.Fatal("Selected must be created")
	}
}

func TestChipGrid_StylesheetMarksSelectedAndDisabled(t *testing.T) {
	out := (&chipgrid.ChipGrid{}).RenderCSS().String()
	for _, st := range []widget.State{widget.Selected, widget.Disabled} {
		if !strings.Contains(out, ".chipgrid__chip["+st.Key()+"=\""+st.Value()+"\"]") {
			t.Errorf("stylesheet lacks the %s chip rule", st)
		}
	}
}

func TestChipGrid_RenderCSS(t *testing.T) {
	sheet := (&chipgrid.ChipGrid{}).RenderCSS()
	if sheet == nil || sheet.String() == "" {
		t.Fatal("empty stylesheet")
	}
}
