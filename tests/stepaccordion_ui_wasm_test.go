//go:build wasm

package components_test

import (
	"syscall/js"
	"testing"

	"webtyp.com/components/stepaccordion"
	"webtyp.com/dom"
	"webtyp.com/html"
	"webtyp.com/widget"
)

func mountAccordion(t *testing.T) (*stepaccordion.StepAccordion, js.Value, *[]string) {
	t.Helper()
	app := js.Global().Get("document").Call("getElementById", "app")
	app.Set("innerHTML", "")
	var reopened []string
	a := &stepaccordion.StepAccordion{}
	a.Steps = []stepaccordion.Step{
		{Key: "one", Title: "One", Body: html.Button().Attr("type", "button").Attr("name", "next-1").Text("next").OnClick(func(dom.Event) { a.Advance("first") })},
		{Key: "two", Title: "Two", Body: html.Button().Attr("type", "button").Attr("name", "next-2").Text("next").OnClick(func(dom.Event) { a.Advance("second") })},
		{Key: "three", Title: "Three", Body: html.P().Text("last")},
	}
	a.OnReopen = func(k string) { reopened = append(reopened, k) }
	a.Init(nil)
	if err := dom.Render("app", a); err != nil {
		t.Fatalf("Render: %v", err)
	}
	return a, app, &reopened
}

func sectionAt(app js.Value, i int) js.Value {
	return app.Call("querySelectorAll", "section.stepaccordion__step").Index(i)
}

func hasState(el js.Value, s widget.State) bool {
	return el.Call("getAttribute", s.Key()).String() == s.Value()
}

func TestWASM_StepAccordion_EditReopensInDOM(t *testing.T) {
	_, app, reopened := mountAccordion(t)
	app.Call("querySelector", "[name='next-1']").Call("click")
	app.Call("querySelector", "[name='next-2']").Call("click")

	if !hasState(sectionAt(app, 0), widget.Locked) || !hasState(sectionAt(app, 2), widget.Open) {
		t.Fatal("after two Advance: step 1 done, step 3 open")
	}

	sectionAt(app, 0).Call("querySelector", ".stepaccordion__edit").Call("click")

	if !hasState(sectionAt(app, 0), widget.Open) {
		t.Error("Edit must reopen step 1")
	}
	if !hasState(sectionAt(app, 1), widget.Disabled) || !hasState(sectionAt(app, 2), widget.Disabled) {
		t.Error("steps after the reopened one must go back to locked")
	}
	if len(*reopened) != 1 || (*reopened)[0] != "one" {
		t.Errorf("OnReopen = %v, want [one]", *reopened)
	}
}

func TestWASM_StepAccordion_LockedStepHasNoFocusableControl(t *testing.T) {
	_, app, _ := mountAccordion(t)
	locked := sectionAt(app, 1)
	if !hasState(locked, widget.Disabled) {
		t.Fatal("step 2 must start locked")
	}
	// What is shown per phase is the stylesheet's job (asserted in
	// TestStepAccordion_StylesheetRevealsByPhase); here: nothing to tab into.
	header := locked.Call("querySelector", ".stepaccordion__header")
	if header.Call("hasAttribute", "tabindex").Bool() {
		t.Error("a locked step's header must not be focusable")
	}
}
