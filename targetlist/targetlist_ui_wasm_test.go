//go:build wasm

package targetlist_test

import (
	"testing"

	"syscall/js"
	"webtyp.com/components/targetlist"
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
// SetItems, not just on a selection change — without that read, reloading
// with a different row count (a filter, a calendar day switch) left the "k /
// N" strip showing a stale total.
func TestMasterCheckCountUpdatesAfterReload(t *testing.T) {
	tl := &targetlist.TargetList{}
	tl.Init(nil)
	tl.SetItems([]targetlist.Item{{ID: "1"}, {ID: "2"}})
	Render("app", tl)

	if got := query(t, ".targetlist__sel-count").Get("textContent").String(); got != "0 / 2" {
		t.Fatalf("count = %q, want 0 / 2", got)
	}

	tl.SetItems([]targetlist.Item{{ID: "1"}, {ID: "2"}, {ID: "3"}})
	if got := query(t, ".targetlist__sel-count").Get("textContent").String(); got != "0 / 3" {
		t.Fatalf("count after reload = %q, want 0 / 3 (itemIDs must resubscribe to t.rows)", got)
	}
}
