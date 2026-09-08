//go:build wasm

package scheduleeditor_test

import (
	"testing"

	"syscall/js"
	"webtyp.com/components/scheduleeditor"
	. "webtyp.com/dom"
)

func TestMain(m *testing.M) {
	app := js.Global().Get("document").Call("createElement", "div")
	app.Set("id", "app")
	js.Global().Get("document").Get("body").Call("appendChild", app)
	m.Run()
}

func query(t *testing.T, sel string) js.Value {
	t.Helper()
	el := js.Global().Get("document").Call("querySelector", sel)
	if el.IsNull() || el.IsUndefined() {
		t.Fatalf("no se encontró %q", sel)
	}
	return el
}

func TestScheduleEditor_AddFormAppearsOnDayPick(t *testing.T) {
	se := &scheduleeditor.ScheduleEditor{
		Pattern: []scheduleeditor.PatternRow{
			{StartMin: 540, EndMin: 1020, Days: []int{1, 2, 3, 4, 5}},
		},
		Exceptions: []scheduleeditor.Exception{
			{ID: "x1", Date: "2026-09-19", Type: scheduleeditor.ExcHoliday},
		},
		Holidays: []string{},
	}
	se.Init(nil)
	Render("app", se)

	day := query(t, "[data-date='2026-09-19']")
	day.Call("click")

	form := query(t, ".scheduleeditor__exc-form")
	open := form.Call("getAttribute", "data-open")
	if open.IsNull() || open.IsUndefined() {
		t.Fatal("expected the exception form to be open after picking a day")
	}

	var got string
	se.OnExceptionAdd = func(ex scheduleeditor.Exception) { got = ex.Date }
	query(t, ".scheduleeditor__exc-add").Call("click")
	if got != "2026-09-19" {
		t.Fatalf("OnExceptionAdd received %q, want 2026-09-19", got)
	}
}

func TestTwoInstancesShareAPage(t *testing.T) {
	se1 := &scheduleeditor.ScheduleEditor{
		Pattern: []scheduleeditor.PatternRow{
			{StartMin: 540, EndMin: 1020, Days: []int{1, 2, 3, 4, 5}},
		},
	}
	se2 := &scheduleeditor.ScheduleEditor{
		Pattern: []scheduleeditor.PatternRow{
			{StartMin: 540, EndMin: 1020, Days: []int{1, 2, 3, 4, 5}},
		},
	}
	se1.Init(nil)
	se2.Init(nil)

	parent := NewElement("div").Child(se1).Child(se2)
	if err := Render("app", parent); err != nil {
		t.Fatalf("mounting two scheduleeditors in one render failed: %v", err)
	}

	editors := js.Global().Get("document").Call("querySelectorAll", ".scheduleeditor")
	if editors.Get("length").Int() != 2 {
		t.Fatalf("expected 2 scheduleeditors in DOM, got %d", editors.Get("length").Int())
	}
}
