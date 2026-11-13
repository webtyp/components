//go:build wasm

package components_test

import (
	"syscall/js"
	"testing"

	"webtyp.com/components/tapinfotip"
	"webtyp.com/dom"
	"webtyp.com/widget"
)

// mountTapInfotip renders the (i) inside a form, the way a field places it:
// the button in a label row, the panel under the control.
func mountTapInfotip(t *testing.T) js.Value {
	t.Helper()
	app := js.Global().Get("document").Call("getElementById", "app")
	app.Set("innerHTML", "")
	tip := &tapinfotip.TapInfotip{Help: "Document number, as written on the document."}
	tip.Init(nil)
	form := dom.NewElement("form").ID("tip-form")
	form.Child(tip.Button(), tip.Panel())
	if err := dom.Render("app", form); err != nil {
		t.Fatalf("Render: %v", err)
	}
	return app
}

// A tap opens the help text and a second tap closes it.
func TestWASM_TapInfotip_TapTogglesPanel(t *testing.T) {
	app := mountTapInfotip(t)
	button := app.Call("querySelector", "button.tapinfotip__button")
	panel := app.Call("querySelector", "p.tapinfotip__panel")

	if hasState(panel, widget.Open) {
		t.Fatal("the help text must be hidden before the first tap")
	}
	button.Call("click")
	if !hasState(panel, widget.Open) || button.Call("getAttribute", "aria-expanded").String() != "true" {
		t.Error("first tap must open the panel and set aria-expanded=true")
	}
	button.Call("click")
	if hasState(panel, widget.Open) || button.Call("getAttribute", "aria-expanded").String() != "false" {
		t.Error("second tap must close the panel and set aria-expanded=false")
	}
}

// The (i) is a button inside a form: a tap must not submit it.
func TestWASM_TapInfotip_TapDoesNotSubmitForm(t *testing.T) {
	app := mountTapInfotip(t)
	submitted := false
	app.Call("querySelector", "form#tip-form").Call("addEventListener", "submit", js.FuncOf(func(this js.Value, args []js.Value) any {
		submitted = true
		return nil
	}))
	app.Call("querySelector", "button.tapinfotip__button").Call("click")
	if submitted {
		t.Error("tapping the (i) submitted the enclosing form")
	}
}
