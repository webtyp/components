//go:build wasm

package selectsearch_test

import (
	"syscall/js"
	"testing"

	"webtyp.com/components/selectsearch"
	. "webtyp.com/dom"
)

func TestMain(m *testing.M) {
	app := js.Global().Get("document").Call("createElement", "div")
	app.Set("id", "app")
	js.Global().Get("document").Get("body").Call("appendChild", app)
	m.Run()
}

func TestTwoInstancesShareAPage(t *testing.T) {
	var sel1, sel2 string
	ss1 := &selectsearch.SelectSearch{
		Options:  []selectsearch.SsOption{{ID: "a", Label: "Option A"}},
		OnSelect: func(id, desc string) { sel1 = id },
	}
	ss2 := &selectsearch.SelectSearch{
		Options:  []selectsearch.SsOption{{ID: "b", Label: "Option B"}},
		OnSelect: func(id, desc string) { sel2 = id },
	}
	ss1.Init(nil)
	ss2.Init(nil)

	parent := NewElement("div").Child(ss1).Child(ss2)
	if err := Render("app", parent); err != nil {
		t.Fatalf("mounting two selectsearches in one render failed: %v", err)
	}

	pickers := js.Global().Get("document").Call("querySelectorAll", ".selectsearch")
	if pickers.Get("length").Int() != 2 {
		t.Fatalf("expected 2 selectsearches in DOM, got %d", pickers.Get("length").Int())
	}

	// Click Option A in picker 1
	item1 := pickers.Call("item", 0).Call("querySelector", "[role='option']")
	if item1.IsNull() || item1.IsUndefined() {
		t.Fatal("option item 1 not found")
	}
	item1.Call("click")

	if sel1 != "a" {
		t.Errorf("picker 1 selection = %q, want 'a'", sel1)
	}
	if sel2 != "" {
		t.Errorf("picker 2 selection = %q, want empty (cross-talk!)", sel2)
	}
}
