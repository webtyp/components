//go:build !wasm

package agedate

import (
	"testing"
)

const testToday = "2026-11-03"

func TestBirthFromAgeYears(t *testing.T) {
	date, ok := birthFromAge("40", testToday)
	if !ok || date != "1986-11-03" {
		t.Fatalf("birthFromAge(40) = %q, %v; want 1986-11-03, true", date, ok)
	}
}

// Under one year there is no age: the exact birth date is required.
func TestAgeUnderOneYearIsRefused(t *testing.T) {
	for _, typed := range []string{"0", "0.5", "0,5"} {
		if _, ok := birthFromAge(typed, testToday); ok {
			t.Errorf("age %q must not make a birth date", typed)
		}
		if _, problem := ageYears(typed); problem != textUnderOneYear {
			t.Errorf("age %q: problem = %q, want %q", typed, problem, textUnderOneYear)
		}
	}
}

func TestAgeMustBeWholeYears(t *testing.T) {
	for _, typed := range []string{"abc", "1e"} {
		if _, problem := ageYears(typed); problem != textWholeYears {
			t.Errorf("age %q: problem = %q, want %q", typed, problem, textWholeYears)
		}
	}
	if _, problem := ageYears(""); problem != "" {
		t.Errorf("nothing typed is not a problem, got %q", problem)
	}
}

func TestAgeFromBirthWholeYears(t *testing.T) {
	cases := map[string]string{
		"1986-11-03": "40", // birthday today
		"1986-11-04": "39", // birthday tomorrow
		"2025-11-03": "1",  // exactly one year
	}
	for born, want := range cases {
		got, ok := ageFromBirth(born, testToday)
		if !ok || got != want {
			t.Errorf("ageFromBirth(%s) = %q, %v; want %q, true", born, got, ok, want)
		}
	}
	if age, ok := ageFromBirth("2026-05-03", testToday); ok {
		t.Errorf("a date under one year old has no age to show, got %q", age)
	}
}

func TestValidate(t *testing.T) {
	a := New()
	if err := a.Validate("2026-02-30"); err == nil {
		t.Fatal("expected error")
	}
	if err := a.Validate("1986-11-03"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// The component only writes to the field the form bound; with nothing bound it is silent.
func TestAlsoSetUsesTheBoundSetter(t *testing.T) {
	a := New()
	a.alsoSet(setEstimated) // nothing bound: must not panic
	var got []string
	a.BindAlsoSets(func(v string) { got = append(got, v) })
	a.alsoSet(setEstimated)
	a.alsoSet(setExact)
	if len(got) != 2 || got[0] != "true" || got[1] != "false" {
		t.Errorf("setter calls = %v, want [true false]", got)
	}
}
