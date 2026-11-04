//go:build !wasm

package targethour

import (
	"strings"
	"testing"

	"webtyp.com/time"
)

func TestTargetHour_RowHasHourLabelAndCheck(t *testing.T) {
	th := &TargetHour{Zone: time.MachineZone()}
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
	z, err := time.LoadZone("America/Santiago")
	if err != nil {
		t.Fatal(err)
	}
	h := hourText(z, "1563240600") // 2019-07-16 01:30 UTC = 21:30 in Santiago (winter)
	want := "21:30"
	if h != want {
		t.Errorf("got %q, want %q", h, want)
	}
}

func TestHourText_PlainText(t *testing.T) {
	h := hourText(time.MachineZone(), "09:30")
	if h != "09:30" {
		t.Errorf("got %q, want %q", h, "09:30")
	}
}
