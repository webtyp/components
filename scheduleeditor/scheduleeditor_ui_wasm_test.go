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
		Week: []scheduleeditor.WeeklyRow{
			{DayOfWeek: 0}, {DayOfWeek: 1, Active: true, WorkStart: 540, WorkFinish: 1020},
			{DayOfWeek: 2}, {DayOfWeek: 3}, {DayOfWeek: 4}, {DayOfWeek: 5}, {DayOfWeek: 6},
		},
		Exceptions: []scheduleeditor.Exception{
			{ID: "x1", Date: "2026-09-19", Type: scheduleeditor.ExcHoliday},
		},
		Holidays: []string{},
	}
	se.Init(nil)
	Render("app", se)

	// The calendar part renders the day for the seeded exception. Tap it.
	day := query(t, "[data-date='2026-09-19']")
	day.Call("click")

	// The inline add-form becomes visible (its container carries data-open).
	form := query(t, ".scheduleeditor__exc-form")
	open := form.Call("getAttribute", "data-open")
	if open.IsNull() || open.IsUndefined() {
		t.Fatal("expected the exception form to be open after picking a day")
	}

	// Submit the add: the callback receives the picked day with an empty ID.
	var got string
	se.OnExceptionAdd = func(ex scheduleeditor.Exception) { got = ex.Date }
	query(t, ".scheduleeditor__exc-add").Call("click")
	if got != "2026-09-19" {
		t.Fatalf("OnExceptionAdd received %q, want 2026-09-19", got)
	}
}
