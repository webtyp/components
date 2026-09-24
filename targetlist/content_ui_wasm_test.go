//go:build wasm

package targetlist_test

import (
	"syscall/js"
	"testing"

	"webtyp.com/components/targetlist"
	. "webtyp.com/dom"
)

// Row-content contract, mirror of targetdate: a reused row must show current
// data, and a click must resolve the current record. The two lists stay
// interchangeable for crudview; a clause that exists in one and not the other
// is how they drift apart.

func mountListRows(t *testing.T, items []targetlist.Item) *targetlist.TargetList {
	t.Helper()
	tl := &targetlist.TargetList{}
	tl.Init(nil)
	tl.SetItems(items)
	if err := Render("app", tl); err != nil {
		t.Fatalf("Render: %v", err)
	}
	return tl
}

func listRow(t *testing.T, key string) js.Value {
	t.Helper()
	el := js.Global().Get("document").Call("querySelector", "[data-row='"+key+"']")
	if el.IsNull() || el.IsUndefined() {
		t.Fatalf("row %s not mounted", key)
	}
	return el
}

func listLabel(row js.Value) string {
	return row.Call("querySelector", ".targetlist__label").Get("textContent").String()
}

func listRows(t *testing.T) []js.Value {
	t.Helper()
	list := query(t, ".targetlist__list")
	kids := list.Get("children")
	out := make([]js.Value, kids.Get("length").Int())
	for i := range out {
		out[i] = kids.Call("item", i)
	}
	return out
}

// Same id, new words (an edit renamed the record): the row keeps its node
// but must show the new label, and a click must ship the new record.
func TestListRowContentUpdatesInPlaceOnSameId(t *testing.T) {
	tl := mountListRows(t, []targetlist.Item{{ID: "1", Label: "a"}})
	ref := listRow(t, "tl-1")

	var got targetlist.Item
	tl.OnSelect = func(it targetlist.Item) { got = it }
	tl.SetItems([]targetlist.Item{{ID: "1", Label: "a2"}})

	again := listRow(t, "tl-1")
	if !again.Equal(ref) {
		t.Error("same id was rebuilt instead of reused")
	}
	if text := listLabel(again); text != "a2" {
		t.Errorf("reused row label = %q, want %q (static build-time text went stale)", text, "a2")
	}
	again.Call("click")
	if got.Label != "a2" {
		t.Errorf("click after update shipped label %q, want %q (stale closure)", got.Label, "a2")
	}
}

// The reported scenario: two records sharing every word but the id, arriving
// through a growth update. Each row keeps its own identity end to end.
func TestListGrowthWithIdenticalLabelsKeepsIdentities(t *testing.T) {
	tl := mountListRows(t, []targetlist.Item{
		{ID: "x", Label: "Same"},
		{ID: "y", Label: "Other"},
	})
	var clicked []string
	tl.OnSelect = func(it targetlist.Item) { clicked = append(clicked, it.ID) }
	tl.SetItems([]targetlist.Item{
		{ID: "z", Label: "Same"},
		{ID: "x", Label: "Same"},
		{ID: "y", Label: "Other"},
	})

	rows := listRows(t)
	if len(rows) != 3 {
		t.Fatalf("3 rows, got %d", len(rows))
	}
	wantLabels := []string{"Same", "Same", "Other"}
	for i, want := range wantLabels {
		if got := listLabel(rows[i]); got != want {
			t.Fatalf("row %d label = %q, want %q", i, got, want)
		}
	}
	for _, row := range rows {
		row.Call("click")
	}
	wantIDs := []string{"z", "x", "y"}
	if len(clicked) != len(wantIDs) {
		t.Fatalf("clicks shipped %v, want %v", clicked, wantIDs)
	}
	for i := range wantIDs {
		if clicked[i] != wantIDs[i] {
			t.Fatalf("click %d shipped id %q, want %q (stale row identity)", i, clicked[i], wantIDs[i])
		}
	}

	tl.Selected.Set("x")
	sel := js.Global().Get("document").Call("querySelectorAll", `[data-selected="true"]`)
	if sel.Get("length").Int() != 1 {
		t.Fatalf("selecting one id must highlight exactly one row, got %d", sel.Get("length").Int())
	}
	if got := sel.Call("item", 0).Call("getAttribute", "data-row").String(); got != "tl-x" {
		t.Errorf("highlighted row = %q, want tl-x", got)
	}
}
