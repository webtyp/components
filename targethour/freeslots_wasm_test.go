//go:build wasm

package targethour_test

import (
	"testing"

	"webtyp.com/components/targethour"
	. "webtyp.com/dom"
)

func TestFreeSlots_ClickCallsOnPickFree(t *testing.T) {
	th := &targethour.TargetHour{}
	th.Init(nil)
	th.FreeSlots = []string{"09:00", "10:30"}

	var picked string
	th.OnPickFree = func(hhmm string) { picked = hhmm }

	th.SetItems([]targethour.Item{})
	Render("app", th)

	el := query(t, "[data-free='10:30']")
	el.Call("click")

	if picked != "10:30" {
		t.Fatalf("OnPickFree received %q, want 10:30", picked)
	}
}
