//go:build wasm

package calendarslider

import (
	"syscall/js"
	"testing"

	. "webtyp.com/dom"
)

// dispatchKey fires a real bubbling KeyboardEvent with the given key on the
// element matching sel, like a user pressing a key while focused there.
func dispatchKey(t *testing.T, sel, key string) {
	t.Helper()
	target := query(t, sel)
	opts := js.Global().Get("Object").New()
	opts.Set("key", key)
	opts.Set("bubbles", true)
	evt := js.Global().Get("KeyboardEvent").New("keydown", opts)
	target.Call("dispatchEvent", evt)
}

// activeIs reports whether document.activeElement is the element matching sel.
func activeIs(sel string) bool {
	active := domDoc().Get("activeElement")
	if active.IsNull() || active.IsUndefined() {
		return false
	}
	want := domDoc().Call("querySelector", sel)
	if want.IsNull() || want.IsUndefined() {
		return false
	}
	return active.Equal(want)
}

// zeroTabsInMonth counts the day buttons carrying tabindex="0" in a month
// card: the roving contract allows exactly one.
func zeroTabsInMonth(month string) int {
	return domDoc().Call("querySelectorAll",
		"[data-month='"+month+"'] .calendarslider__day-button[tabindex='0']").Get("length").Int()
}

// TestArrowKeysMoveFocus is the assertion that failed in the running demo
// before this change: focus did not move on ArrowRight, ArrowDown or Home,
// because dom.Event could not read a key press at all. The handler now goes
// through dom.OnKeyDown and dom.Key.
func TestArrowKeysMoveFocus(t *testing.T) {
	c := &CalendarSlider{
		Start: "2026-08",
		Occupation: []OccupationDay{
			{Date: "2026-08-11", Percent: 60},
			{Date: "2026-08-12", Percent: 30},
			{Date: "2026-08-13", Percent: 90},
		},
	}
	c.Init(nil)
	Render("app", c.Render())

	// one tab stop per month card, on the first selectable day (no Selected,
	// today is not in August 2026)
	if n := zeroTabsInMonth("2026-08"); n != 1 {
		t.Fatalf("want exactly one tabindex=0 in August, got %d", n)
	}
	first := "[data-date='2026-08-11'] .calendarslider__day-button"
	if tab := query(t, first).Call("getAttribute", "tabindex").String(); tab != "0" {
		t.Fatalf("first selectable day tabindex = %q, want 0", tab)
	}

	// focus it, like Tab landing on the card
	query(t, first).Call("focus")
	if !activeIs(first) {
		t.Fatal("could not focus the first day button")
	}

	// ArrowRight: focus moves to the next day, the roving stop follows
	dispatchKey(t, first, "ArrowRight")
	second := "[data-date='2026-08-12'] .calendarslider__day-button"
	if !activeIs(second) {
		t.Error("after ArrowRight focus should be on 2026-08-12")
	}
	if n := zeroTabsInMonth("2026-08"); n != 1 {
		t.Errorf("want exactly one tabindex=0 after ArrowRight, got %d", n)
	}
	if tab := query(t, second).Call("getAttribute", "tabindex").String(); tab != "0" {
		t.Errorf("focused day tabindex = %q, want 0", tab)
	}

	// Home: back to the week's first selectable day
	dispatchKey(t, second, "Home")
	if !activeIs(first) {
		t.Error("after Home focus should be back on 2026-08-11")
	}

	// End: to the week's last selectable day
	dispatchKey(t, first, "End")
	third := "[data-date='2026-08-13'] .calendarslider__day-button"
	if !activeIs(third) {
		t.Error("after End focus should be on 2026-08-13")
	}

	// ArrowRight past the last selectable day of the week goes nowhere:
	// movement stays inside the month card and never wraps the strip.
	dispatchKey(t, third, "ArrowRight")
	if !activeIs(third) {
		t.Error("ArrowRight at the row end must not move focus")
	}
}
