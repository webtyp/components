//go:build wasm

package targetdate_test

import (
	"testing"

	"syscall/js"
	"webtyp.com/components/targetdate"
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

// TestMasterCheckCountUpdatesAfterReload is the live-DOM half of the fix a
// //go:build !wasm test cannot cover: Render().String() always re-evaluates
// its bound closures fresh, so it never notices a missing subscription — only
// a widget mounted ONCE and then mutated, like a real page, does. itemIDs()
// must read t.rows so the header's bound count text resubscribes on every
// SetItems, not just on a selection change — without that read, switching a
// calendar's day left the "k / N" strip showing the previous day's total.
func TestMasterCheckCountUpdatesAfterReload(t *testing.T) {
	td := &targetdate.TargetDate{}
	td.Init(nil)
	td.SetItems([]targetdate.Item{{ID: "1"}, {ID: "2"}})
	Render("app", td)

	if got := query(t, ".targetdate__sel-count").Get("textContent").String(); got != "0 / 2" {
		t.Fatalf("count = %q, want 0 / 2", got)
	}

	td.SetItems([]targetdate.Item{{ID: "1"}, {ID: "2"}, {ID: "3"}})
	if got := query(t, ".targetdate__sel-count").Get("textContent").String(); got != "0 / 3" {
		t.Fatalf("count after reload = %q, want 0 / 3 (itemIDs must resubscribe to t.rows)", got)
	}
}
