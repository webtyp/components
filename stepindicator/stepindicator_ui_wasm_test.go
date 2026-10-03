//go:build wasm

package stepindicator_test

import (
	"syscall/js"
	"testing"

	"webtyp.com/components/stepindicator"
	"webtyp.com/dom"
)

func TestMain(m *testing.M) {
	doc := js.Global().Get("document")
	app := doc.Call("getElementById", "app")
	if app.IsNull() || app.IsUndefined() {
		app = doc.Call("createElement", "div")
		app.Set("id", "app")
		doc.Get("body").Call("appendChild", app)
	}
	m.Run()
}

func TestStepIndicator_Click(t *testing.T) {
	app := js.Global().Get("document").Call("getElementById", "app")
	app.Set("innerHTML", "")

	var clickedIdx int = -1
	comp := &stepindicator.StepIndicator{
		Steps: []stepindicator.Step{
			{Key: "s1", Label: "Step 1"},
			{Key: "s2", Label: "Step 2"},
		},
		Active: 0,
		OnChange: func(idx int) {
			clickedIdx = idx
		},
	}
	comp.Init(nil)
	dom.Render("app", comp)

	buttons := app.Call("querySelectorAll", "button")
	if buttons.Length() < 2 {
		t.Fatalf("expected at least 2 step buttons, got %d", buttons.Length())
	}

	buttons.Index(1).Call("click")
	if clickedIdx != 1 {
		t.Fatalf("expected clickedIdx to be 1, got %d", clickedIdx)
	}
}
