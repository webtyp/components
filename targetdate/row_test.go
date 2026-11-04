//go:build !wasm

package targetdate

import (
	"strconv"
	"strings"
	"testing"

	"webtyp.com/time"
)

// Internal, so it can reach buildRow: Render() binds its children and SSR does
// not serialize a children binding, so this is the only place a row's markup
// can actually be asserted. targetlist carries the same test — the two
// components must stay interchangeable for crudview, and a check that exists
// in one and not the other is exactly how they drift apart.
func TestTargetDate_RowHasLeadLabelAndCheck(t *testing.T) {
	td := &TargetDate{Zone: time.MachineZone()}
	td.Init(nil)

	html := td.buildRow(Item{
		ID: "7", Label: "Fractura", Description: "dr. Tony Stark",
		Badge: "Hoy", // Plain text badge; will render "Hoy" in the main line.
	}).String()

	for _, want := range []string{"targetdate__row", "Fractura", "targetdate__sel-check", "Hoy"} {
		if !strings.Contains(html, want) {
			t.Errorf("buildRow output missing %q\ngot: %s", want, html)
		}
	}
	for _, unwanted := range []string{"targetdate__button", "targetdate__options", "targetdate__item-danger", "Eliminar"} {
		if strings.Contains(html, unwanted) {
			t.Errorf("buildRow must not render per-row menu artifact %q\ngot: %s", unwanted, html)
		}
	}
}

func TestLeadLines_StoredDate(t *testing.T) {
	u := time.DateUTC(2026, 10, 27, 12, 0, 0)
	badge := strconv.FormatInt(int64(u), 10)
	top, main, bottom := leadLines(santiago(t), badge)
	if top != "Tue" { // we aren't bootstrapping language translations in test, so fallback to English keys
		t.Errorf("got top %q, want %q", top, "Tue")
	}
	if main != "27" {
		t.Errorf("got main %q, want %q", main, "27")
	}
	if bottom != "Oct" {
		t.Errorf("got bottom %q, want %q", bottom, "Oct")
	}
}

func TestLeadLines_PlainText(t *testing.T) {
	top, main, bottom := leadLines(santiago(t), "Hoy")
	if top != "" || main != "Hoy" || bottom != "" {
		t.Errorf("got %q, %q, %q, want %q, %q, %q", top, main, bottom, "", "Hoy", "")
	}
}

func TestLeadLines_Empty(t *testing.T) {
	top, main, bottom := leadLines(santiago(t), "")
	if top != "" || main != "" || bottom != "" {
		t.Errorf("got %q, %q, %q, want empty", top, main, bottom)
	}
}

func santiago(t *testing.T) time.Zone {
	z, err := time.LoadZone("America/Santiago")
	if err != nil {
		t.Fatal(err)
	}
	return z
}

// A winter night booking: 2019-07-16 01:30 UTC is Monday the 15th, 21:30, in Santiago (UTC is Tuesday the 16th).
func TestLeadLines_ZoneNotUTC(t *testing.T) {
	top, main, bottom := leadLines(santiago(t), "1563240600")
	if top != "Mon" || main != "15" || bottom != "Jul" {
		t.Errorf("got %q %q %q, want Mon 15 Jul", top, main, bottom)
	}
}
