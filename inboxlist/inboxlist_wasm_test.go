//go:build wasm

package inboxlist_test

import (
	"syscall/js"
	"testing"

	"webtyp.com/components/inboxlist"
	. "webtyp.com/dom"
)

func TestMain(m *testing.M) {
	app := js.Global().Get("document").Call("createElement", "div")
	app.Set("id", "app")
	js.Global().Get("document").Get("body").Call("appendChild", app)
	m.Run()
}

func TestInboxList_SelectWASM(t *testing.T) {
	var selectedID string
	l := &inboxlist.InboxList{
		OnSelect: func(id string) {
			selectedID = id
		},
	}

	l.SetRows([]inboxlist.Row{
		{ID: "msg1", Title: "Alice", Preview: "Hi", Time: "12:00", Unread: 1},
		{ID: "msg2", Title: "Bob", Preview: "Hey", Time: "12:01", Unread: 0},
	})

	if err := Render("app", l); err != nil {
		t.Fatalf("mounting inboxlist failed: %v", err)
	}

	doc := js.Global().Get("document")
	row1 := doc.Call("querySelector", "[role='option']")
	if row1.IsNull() || row1.IsUndefined() {
		t.Fatal("expected row element in DOM")
	}

	row1.Call("click")

	if selectedID != "msg1" {
		t.Errorf("expected selectedID 'msg1', got %q", selectedID)
	}
	if l.Selected.Get() != "msg1" {
		t.Errorf("expected Selected signal 'msg1', got %q", l.Selected.Get())
	}
}
