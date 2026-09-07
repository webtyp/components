//go:build !wasm

package targethour

import (
	"strings"
	"testing"
)

// Free slots render as distinct light rows AFTER the booked-item rows.
func TestFreeSlots_RenderedAfterItems(t *testing.T) {
	th := &TargetHour{}
	th.Init(nil)
	th.FreeSlots = []string{"09:00", "10:30", "14:00"}
	th.SetItems([]Item{
		{ID: "1", LeadMain: "08:00", Label: "a"},
		{ID: "2", LeadMain: "08:30", Label: "b"},
	})

	nodes := th.rows.Get()
	if len(nodes) != 5 {
		t.Fatalf("rows = %d, want 5 (2 items + 3 free slots)", len(nodes))
	}
	// The last three are free-slot rows, in order.
	for i, want := range []string{"09:00", "10:30", "14:00"} {
		html := nodes[2+i].String()
		if !strings.Contains(html, "targethour__free") || !strings.Contains(html, want) {
			t.Errorf("free row %d missing class/hour %q:\n%s", i, want, html)
		}
	}
	// Item rows carry the free class? They must NOT.
	for _, it := range nodes[:2] {
		if strings.Contains(it.String(), "targethour__free") {
			t.Errorf("item row must not carry the free class:\n%s", it.String())
		}
	}
}

// FreeSlots nil must leave the HTML identical to today (no free rows).
func TestFreeSlots_NilNoRegression(t *testing.T) {
	th := &TargetHour{}
	th.Init(nil)
	th.SetItems([]Item{{ID: "1", LeadMain: "08:00", Label: "a"}})

	nodes := th.rows.Get()
	if len(nodes) != 1 {
		t.Fatalf("rows = %d, want 1 (no free slots)", len(nodes))
	}
	if strings.Contains(nodes[0].String(), "targethour__free") {
		t.Errorf("FreeSlots nil must not render free rows:\n%s", nodes[0].String())
	}
}

// buildFreeSlot emits the clickable row with the hour and the marker.
func TestFreeSlots_BuildRow(t *testing.T) {
	th := &TargetHour{}
	th.Init(nil)

	html := th.buildFreeSlot("11:15").String()

	for _, want := range []string{"targethour__free", "11:15", "targethour__free-add", "+"} {
		if !strings.Contains(html, want) {
			t.Errorf("buildFreeSlot missing %q:\n%s", want, html)
		}
	}
	for _, unwanted := range []string{"targethour__sel-check", "targethour__badge"} {
		if strings.Contains(html, unwanted) {
			t.Errorf("buildFreeSlot must not render item chrome %q:\n%s", unwanted, html)
		}
	}
}
