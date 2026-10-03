//go:build wasm

package segmentedcontrol_test

import (
	"syscall/js"
	"testing"

	"webtyp.com/components/segmentedcontrol"
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

func TestSegmentedControl_Click(t *testing.T) {
	app := js.Global().Get("document").Call("getElementById", "app")
	app.Set("innerHTML", "")

	var selected string
	comp := &segmentedcontrol.SegmentedControl{
		Options: []segmentedcontrol.Option{
			{Value: "A", Label: "A"},
			{Value: "B", Label: "B"},
		},
		Selected: "A",
		OnChange: func(val string) {
			selected = val
		},
	}
	comp.Init(nil)
	dom.Render("app", comp)

	buttons := app.Call("querySelectorAll", "button")
	if buttons.Length() < 2 {
		t.Fatalf("expected at least 2 buttons, got %d", buttons.Length())
	}

	buttons.Index(1).Call("click")
	if selected != "B" {
		t.Fatalf("expected selected to be 'B', got %q", selected)
	}
}
