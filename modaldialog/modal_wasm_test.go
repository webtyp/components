//go:build wasm

package modaldialog

import (
	"testing"

	"syscall/js"
	. "webtyp.com/dom"
)

func TestMain(m *testing.M) {
	app := js.Global().Get("document").Call("createElement", "div")
	app.Set("id", "app")
	js.Global().Get("document").Get("body").Call("appendChild", app)
	m.Run()
}

// TestModalClickContract pins the restructure of PLAN.md Stage 4: there is no
// click-catcher sibling anymore — the root IS the wash and the click target
// (a click on it closes the dialog), and the panel stops the click from
// reaching it (interacting with the dialog never closes it). The panel wins
// by being the wash's own in-flow child, so the click model needs no stacking.
func TestModalClickContract(t *testing.T) {
	m := &ModalDialog{
		Title:   "Confirm",
		Content: NewElement("button").ID("dlg-ok").Text("OK"),
	}
	m.Init(nil)
	root := m.Render()
	Render("app", root)

	doc := js.Global().Get("document")
	visible := func() bool {
		dl := doc.Call("querySelector", ".modaldialog")
		return dl.Get("parentElement").Get("style").Get("display").String() != "none"
	}

	if visible() {
		t.Fatal("dialog must be hidden initially")
	}

	m.Open()
	if !visible() {
		t.Fatal("dialog must be visible after Open()")
	}

	// Clicking the wash (the root itself) closes the dialog.
	doc.Call("querySelector", ".modaldialog").Call("click")
	if visible() {
		t.Error("click on the wash must close the dialog")
	}

	// Clicking inside the panel must not close it.
	m.Open()
	doc.Call("querySelector", ".modaldialog__panel").Call("click")
	if !visible() {
		t.Error("click on the panel must not close the dialog")
	}

	// Clicking a control inside the panel must not close it either.
	m.Open()
	doc.Call("querySelector", "#dlg-ok").Call("click")
	if !visible() {
		t.Error("click on panel content must not close the dialog")
	}
}
