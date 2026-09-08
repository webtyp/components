//go:build wasm

package listselect_test

import (
	"testing"

	"syscall/js"
	"webtyp.com/components/listselect"
	. "webtyp.com/dom"
	"webtyp.com/widget"
)

const uiName = widget.Name("uitest")

var ids = func() []string { return []string{"a", "b", "c"} }

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

func TestHeaderSelectsAllAndClears(t *testing.T) {
	var m listselect.Mode
	Render("app", listselect.Header(&m, ids, uiName))

	if got := query(t, ".uitest__sel-count").Get("textContent").String(); got != "0 / 3" {
		t.Fatalf("count = %q, want 0 / 3", got)
	}

	query(t, ".uitest__sel-all").Call("click")
	if m.Count() != 3 {
		t.Fatalf("Count = %d, want 3 after select-all", m.Count())
	}
	if got := query(t, ".uitest__sel-count").Get("textContent").String(); got != "3 / 3" {
		t.Fatalf("count = %q, want 3 / 3", got)
	}

	query(t, ".uitest__sel-all").Call("click")
	if m.Count() != 0 {
		t.Fatalf("Count = %d, want 0 after deselect-all", m.Count())
	}
	if got := query(t, ".uitest__sel-count").Get("textContent").String(); got != "0 / 3" {
		t.Fatalf("count = %q, want 0 / 3 after deselect-all", got)
	}
}

// TestHeaderRestsUntilSomethingChecked: entering selection mode alone must
// NOT tone the box — that made tapping select-all invisible, since the box
// already looked "active" at zero. Only a mark (Toggle/CheckAll) does.
func stateAttr(el js.Value, name string) string {
	v := el.Call("getAttribute", name)
	if v.IsNull() || v.IsUndefined() {
		return ""
	}
	return v.String()
}

func TestHeaderRestsUntilSomethingChecked(t *testing.T) {
	var m listselect.Mode
	m.SetOn(true)
	Render("app", listselect.Header(&m, ids, uiName))

	box := query(t, ".uitest__sel-all")
	if v := stateAttr(box, "data-invalid"); v != "" {
		t.Errorf("an empty selection must not carry data-invalid, got %q", v)
	}
	if v := stateAttr(box, "data-selected"); v != "" {
		t.Errorf("an empty selection must not carry data-selected, got %q", v)
	}

	m.Toggle("a")
	if v := stateAttr(box, "data-selected"); v != "true" {
		t.Errorf("marking a row must put data-selected on the box, got %q", v)
	}
}

// TestHeaderDangerArmsInvalidTone: the box's OWN Invalid/Selected state (not
// its glyph — the glyph is fixed, see selectall) tones the background solid
// once something is marked: Danger red while marked for delete, Accent
// amber while marked for a bulk edit. Under the armed danger tone the box
// must carry data-invalid, never data-selected; once disarmed it flips to
// data-selected.
func TestHeaderDangerArmsInvalidTone(t *testing.T) {
	var m listselect.Mode
	m.SetDanger(true)
	m.SetOn(true)
	m.Toggle("a")
	Render("app", listselect.Header(&m, ids, uiName))

	box := query(t, ".uitest__sel-all")
	if v := stateAttr(box, "data-invalid"); v != "true" {
		t.Errorf("danger tone must put data-invalid on the select-all box, got %q", v)
	}
	if v := stateAttr(box, "data-selected"); v != "" {
		t.Errorf("danger tone must NOT put data-selected on the select-all box, got %q", v)
	}

	m.SetDanger(false)
	if v := stateAttr(box, "data-invalid"); v != "" {
		t.Errorf("disarmed tone must drop data-invalid, got %q", v)
	}
	if v := stateAttr(box, "data-selected"); v != "true" {
		t.Errorf("disarmed tone must put data-selected, got %q", v)
	}
}

func TestRowOfEditAndDanger(t *testing.T) {
	var m listselect.Mode
	m.SetOn(true)
	m.Toggle("b")

	r := listselect.RowOf(&m, "b", uiName)
	if !r.Edit.Get() {
		t.Error("Edit must be true for the marked row without the danger tone")
	}
	if r.Danger.Get() {
		t.Error("Danger must be false for the marked row without the danger tone")
	}

	m.SetDanger(true)
	if r.Edit.Get() {
		t.Error("Edit must be false once the danger tone is armed")
	}
	if !r.Danger.Get() {
		t.Error("Danger must be true once the danger tone is armed")
	}
}
