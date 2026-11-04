//go:build !wasm

package agedate

import (
	"testing"
	"webtyp.com/time"
)

func TestBirthFromAgeYears(t *testing.T) {
	today := time.Date(2026, 11, 3, 0, 0, 0)
	date, ok := birthFromAge("40", today)
	if !ok {
		t.Fatal("expected ok")
	}
	if date != "1986-11-03" {
		t.Fatalf("expected 1986-11-03, got %s", date)
	}
}

func TestBirthFromAgeMonths(t *testing.T) {
	today := time.Date(2026, 11, 3, 0, 0, 0)
	date, ok := birthFromAge("0.5", today)
	if !ok {
		t.Fatal("expected ok")
	}
	if date != "2026-05-03" {
		t.Fatalf("expected 2026-05-03, got %s", date)
	}

	date, ok = birthFromAge("0,5", today)
	if !ok {
		t.Fatal("expected ok")
	}
	if date != "2026-05-03" {
		t.Fatalf("expected 2026-05-03, got %s", date)
	}
}

func TestBirthFromAgeEmpty(t *testing.T) {
	today := time.Date(2026, 11, 3, 0, 0, 0)
	_, ok := birthFromAge("", today)
	if ok {
		t.Fatal("expected !ok")
	}
	_, ok = birthFromAge("0", today)
	if ok {
		t.Fatal("expected !ok")
	}
}

func TestAgeFromBirth(t *testing.T) {
	today := time.Date(2026, 11, 3, 0, 0, 0)
	age, ok := ageFromBirth("1986-11-03", today)
	if !ok {
		t.Fatal("expected ok")
	}
	if age != "40.0" {
		t.Fatalf("expected 40.0, got %s", age)
	}

	age, ok = ageFromBirth("2026-05-03", today)
	if !ok {
		t.Fatal("expected ok")
	}
	if age != "0.5" {
		t.Fatalf("expected 0.5, got %s", age)
	}
}

func TestValidate(t *testing.T) {
	a := New()
	err := a.Validate("2026-02-30")
	if err == nil {
		t.Fatal("expected error")
	}

	err = a.Validate("1986-11-03")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
