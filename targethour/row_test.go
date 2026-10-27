//go:build !wasm

package targethour

import (
	"strconv"
	"strings"
	"testing"

	"webtyp.com/time"
)

func TestTargetHour_RowHasHourLabelAndCheck(t *testing.T) {
	th := &TargetHour{}
	th.Init(nil)

	html := th.buildRow(Item{
		ID: "7", Label: "Fractura", Description: "dr. Tony Stark",
		Badge: "10:30",
	}).String()

	for _, want := range []string{"targethour__row", "Fractura", "targethour__sel-check", "targethour__hour", "10:30"} {
		if !strings.Contains(html, want) {
			t.Errorf("buildRow output missing %q\ngot: %s", want, html)
		}
	}
	for _, unwanted := range []string{"targethour__button", "targethour__options", "targethour__item-danger", "Eliminar"} {
		if strings.Contains(html, unwanted) {
			t.Errorf("buildRow must not render per-row menu artifact %q\ngot: %s", unwanted, html)
		}
	}
}

func TestHourText_StoredInstant(t *testing.T) {
	// A time at 09:30 UTC
	u := time.Date(2026, 10, 27, 9, 30, 0)
	badge := strconv.FormatInt(int64(u), 10)
	h := hourText(badge)
	if h != "09:30" {
		t.Errorf("got %q, want %q", h, "09:30")
	}
}

func TestHourText_PlainText(t *testing.T) {
	h := hourText("09:30")
	if h != "09:30" {
		t.Errorf("got %q, want %q", h, "09:30")
	}
}
