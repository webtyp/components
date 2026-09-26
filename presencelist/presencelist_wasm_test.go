//go:build wasm

package presencelist_test

import (
	"syscall/js"
	"testing"

	"webtyp.com/components/presencelist"
	. "webtyp.com/dom"
)

func TestMain(m *testing.M) {
	app := js.Global().Get("document").Call("createElement", "div")
	app.Set("id", "app")
	js.Global().Get("document").Get("body").Call("appendChild", app)
	m.Run()
}

func TestPresenceList_SelectWASM(t *testing.T) {
	var selectedID string
	pl := &presencelist.PresenceList{
		OnSelect: func(id string) {
			selectedID = id
		},
	}

	pl.SetPeople([]presencelist.Person{
		{ID: "p1", Label: "Alice", Online: true},
		{ID: "p2", Label: "Bob", Online: false},
	})

	if err := Render("app", pl); err != nil {
		t.Fatalf("mounting presencelist failed: %v", err)
	}

	doc := js.Global().Get("document")
	row1 := doc.Call("querySelector", "[role='option']")
	if row1.IsNull() || row1.IsUndefined() {
		t.Fatal("expected row element in DOM")
	}

	row1.Call("click")

	if selectedID != "p1" {
		t.Errorf("expected selectedID 'p1', got %q", selectedID)
	}
}
