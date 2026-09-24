//go:build wasm

package targetdate_test

import (
	"syscall/js"
	"testing"

	"webtyp.com/components/targetdate"
	. "webtyp.com/dom"
)

// Row-content contract: a reused row must show current data, and a click
// must resolve the current record — never the strings the node was built
// with. Static Text() at build time froze the first render into nodes the
// reconciler now correctly reuses, so same id + new words (an edit) and the
// reported duplicate-highlight scenario need live-DOM proof here.

func mountDateRows(t *testing.T, items []targetdate.Item) *targetdate.TargetDate {
	t.Helper()
	td := &targetdate.TargetDate{}
	td.Init(nil)
	td.SetItems(items)
	if err := Render("app", td); err != nil {
		t.Fatalf("Render: %v", err)
	}
	return td
}

func dateRow(t *testing.T, key string) js.Value {
	t.Helper()
	el := js.Global().Get("document").Call("querySelector", "[data-row='"+key+"']")
	if el.IsNull() || el.IsUndefined() {
		t.Fatalf("row %s not mounted", key)
	}
	return el
}

func dateLabel(row js.Value) string {
	return row.Call("querySelector", ".targetdate__label").Get("textContent").String()
}

func dateRows(t *testing.T) []js.Value {
	t.Helper()
	list := query(t, ".targetdate__list")
	kids := list.Get("children")
	out := make([]js.Value, kids.Get("length").Int())
	for i := range out {
		out[i] = kids.Call("item", i)
	}
	return out
}

// Same id, new words (an edit renamed the record): the row keeps its node
// but must show the new label, and a click must ship the new record.
func TestDateRowContentUpdatesInPlaceOnSameId(t *testing.T) {
	td := mountDateRows(t, []targetdate.Item{{ID: "1", Label: "a"}})
	ref := dateRow(t, "td-1")

	var got targetdate.Item
	td.OnSelect = func(it targetdate.Item) { got = it }
	td.SetItems([]targetdate.Item{{ID: "1", Label: "a2"}})

	again := dateRow(t, "td-1")
	if !again.Equal(ref) {
		t.Error("same id was rebuilt instead of reused")
	}
	if text := dateLabel(again); text != "a2" {
		t.Errorf("reused row label = %q, want %q (static build-time text went stale)", text, "a2")
	}
	again.Call("click")
	if got.Label != "a2" {
		t.Errorf("click after update shipped label %q, want %q (stale closure)", got.Label, "a2")
	}
}

// The reported scenario: two records sharing every word but the id, arriving
// through a growth update. Each row keeps its own identity end to end.
func TestDateGrowthWithIdenticalLabelsKeepsIdentities(t *testing.T) {
	td := mountDateRows(t, []targetdate.Item{
		{ID: "x", Label: "Same"},
		{ID: "y", Label: "Other"},
	})
	var clicked []string
	td.OnSelect = func(it targetdate.Item) { clicked = append(clicked, it.ID) }
	td.SetItems([]targetdate.Item{
		{ID: "z", Label: "Same"},
		{ID: "x", Label: "Same"},
		{ID: "y", Label: "Other"},
	})

	rows := dateRows(t)
	if len(rows) != 3 {
		t.Fatalf("3 rows, got %d", len(rows))
	}
	wantLabels := []string{"Same", "Same", "Other"}
	for i, want := range wantLabels {
		if got := dateLabel(rows[i]); got != want {
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

	td.Selected.Set("x")
	sel := js.Global().Get("document").Call("querySelectorAll", `[data-selected="true"]`)
	if sel.Get("length").Int() != 1 {
		t.Fatalf("selecting one id must highlight exactly one row, got %d", sel.Get("length").Int())
	}
	if got := sel.Call("item", 0).Call("getAttribute", "data-row").String(); got != "td-x" {
		t.Errorf("highlighted row = %q, want td-x", got)
	}
}
