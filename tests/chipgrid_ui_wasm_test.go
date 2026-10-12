//go:build wasm

package components_test

import (
	"syscall/js"
	"testing"

	"webtyp.com/components/chipgrid"
	"webtyp.com/dom"
	"webtyp.com/widget"
)

func mountGrid(t *testing.T, chips []chipgrid.Chip, withCallback bool) (*chipgrid.ChipGrid, js.Value, *[]string) {
	t.Helper()
	app := js.Global().Get("document").Call("getElementById", "app")
	app.Set("innerHTML", "")
	var picked []string
	g := &chipgrid.ChipGrid{}
	if withCallback {
		g.OnPick = func(v string) { picked = append(picked, v) }
	}
	g.Init(nil)
	g.SetChips(chips)
	if err := dom.Render("app", g); err != nil {
		t.Fatalf("Render: %v", err)
	}
	return g, app, &picked
}

func chipAt(app js.Value, i int) js.Value {
	return app.Call("querySelectorAll", ".chipgrid__chip").Index(i)
}

func pressKey(el js.Value, key string) {
	init := js.Global().Get("Object").New()
	init.Set("key", key)
	init.Set("bubbles", true)
	el.Call("dispatchEvent", js.Global().Get("KeyboardEvent").New("keydown", init))
}

func activeValue() string {
	return js.Global().Get("document").Get("activeElement").Call("getAttribute", "data-value").String()
}

func hourChips() []chipgrid.Chip {
	return []chipgrid.Chip{
		{Value: "09:00", Label: "09:00"},
		{Value: "09:30", Label: "09:30", Disabled: true},
		{Value: "10:00", Label: "10:00"},
	}
}

func fiveChips() []chipgrid.Chip {
	return []chipgrid.Chip{
		{Value: "a", Label: "a"},
		{Value: "b", Label: "b", Disabled: true},
		{Value: "c", Label: "c"},
		{Value: "d", Label: "d"},
		{Value: "e", Label: "e", Disabled: true},
	}
}

func TestWASM_ChipGrid_ClickPicksAndCallsOnPick(t *testing.T) {
	g, app, picked := mountGrid(t, hourChips(), true)
	chipAt(app, 2).Call("click")
	if g.Selected.Get() != "10:00" {
		t.Errorf("Selected = %q, want 10:00", g.Selected.Get())
	}
	if len(*picked) != 1 || (*picked)[0] != "10:00" {
		t.Errorf("OnPick = %v, want [10:00]", *picked)
	}
	if !hasState(chipAt(app, 2), widget.Selected) {
		t.Error("the picked chip is marked Selected in the DOM")
	}
}

func TestWASM_ChipGrid_DisabledClickIgnored(t *testing.T) {
	g, app, picked := mountGrid(t, hourChips(), true)
	chipAt(app, 1).Call("click")
	if g.Selected.Get() != "" || len(*picked) != 0 {
		t.Errorf("a disabled chip must not pick: Selected=%q OnPick=%v", g.Selected.Get(), *picked)
	}
}

func TestWASM_ChipGrid_ClickWithoutCallback(t *testing.T) {
	g, app, _ := mountGrid(t, hourChips(), false)
	chipAt(app, 0).Call("click")
	if g.Selected.Get() != "09:00" {
		t.Errorf("Selected = %q, want 09:00", g.Selected.Get())
	}
}

func TestWASM_ChipGrid_ArrowsSkipDisabled(t *testing.T) {
	_, app, _ := mountGrid(t, fiveChips(), true)
	chipAt(app, 0).Call("focus")
	pressKey(chipAt(app, 0), "ArrowRight")
	if activeValue() != "c" {
		t.Fatalf("ArrowRight from a: focus on %q, want c (b is disabled)", activeValue())
	}
	pressKey(chipAt(app, 2), "ArrowDown")
	if activeValue() != "d" {
		t.Fatalf("ArrowDown from c: focus on %q, want d", activeValue())
	}
	pressKey(chipAt(app, 3), "ArrowUp")
	pressKey(chipAt(app, 2), "ArrowLeft")
	if activeValue() != "a" {
		t.Fatalf("ArrowUp then ArrowLeft from d: focus on %q, want a", activeValue())
	}
	if chipAt(app, 0).Call("getAttribute", "tabindex").String() != "0" || chipAt(app, 2).Call("getAttribute", "tabindex").String() != "-1" {
		t.Error("the roving tab stop follows the focus")
	}
}

func TestWASM_ChipGrid_HomeEnd(t *testing.T) {
	_, app, _ := mountGrid(t, fiveChips(), true)
	chipAt(app, 2).Call("focus")
	pressKey(chipAt(app, 2), "End")
	if activeValue() != "d" {
		t.Fatalf("End: focus on %q, want d (e is disabled)", activeValue())
	}
	pressKey(chipAt(app, 3), "Home")
	if activeValue() != "a" {
		t.Fatalf("Home: focus on %q, want a", activeValue())
	}
}

func TestWASM_ChipGrid_ArrowAtEdgeStays(t *testing.T) {
	_, app, _ := mountGrid(t, fiveChips(), true)
	chipAt(app, 3).Call("focus")
	pressKey(chipAt(app, 3), "ArrowRight") // only a disabled chip after d
	if activeValue() != "d" {
		t.Errorf("focus moved to %q, want it to stay on d", activeValue())
	}
	pressKey(chipAt(app, 3), "x") // not a navigation key
	if activeValue() != "d" {
		t.Errorf("an unrelated key moved focus to %q", activeValue())
	}
}

func TestWASM_ChipGrid_EnterAndSpacePick(t *testing.T) {
	g, app, picked := mountGrid(t, fiveChips(), true)
	chipAt(app, 2).Call("focus")
	pressKey(chipAt(app, 2), "Enter")
	if g.Selected.Get() != "c" {
		t.Fatalf("Enter: Selected = %q, want c", g.Selected.Get())
	}
	pressKey(chipAt(app, 3), " ")
	if g.Selected.Get() != "d" || len(*picked) != 2 {
		t.Errorf("Space: Selected = %q, OnPick = %v", g.Selected.Get(), *picked)
	}
}
