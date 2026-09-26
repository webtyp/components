//go:build !wasm

package presencelist

import (
	"strings"
	"testing"
)

func TestPresenceList_IdempotentAndRootClass(t *testing.T) {
	l := &PresenceList{}
	s1 := l.Render().String()
	s2 := l.Render().String()
	if s1 != s2 {
		t.Fatalf("expected idempotent Render(), got %q vs %q", s1, s2)
	}
	if !strings.Contains(s1, "presencelist") {
		t.Fatalf("expected output to contain root class presencelist, got %q", s1)
	}
}

func TestPresenceList_RenderCSS(t *testing.T) {
	l := &PresenceList{}
	sheet := l.RenderCSS()
	if sheet == nil {
		t.Fatal("RenderCSS returned nil")
	}
}

func TestPresenceList_Sorting(t *testing.T) {
	l := &PresenceList{
		OnlineLabel:  "En línea",
		OfflineLabel: "Desconectado",
	}

	l.SetPeople([]Person{
		{ID: "b", Label: "Bob", Online: false},
		{ID: "a", Label: "Alice", Online: true},
		{ID: "c", Label: "Charlie", Online: true},
	})

	people := l.People()
	if len(people) != 3 {
		t.Fatalf("expected 3 people, got %d", len(people))
	}

	// Order must be Alice (online), Charlie (online), Bob (offline)
	expectedIDs := []string{"a", "c", "b"}
	for i, wantID := range expectedIDs {
		if people[i].ID != wantID {
			t.Errorf("person %d: got ID %q, want %q", i, people[i].ID, wantID)
		}
	}

	rowA := l.buildRow(people[0]).String()
	rowB := l.buildRow(people[2]).String()

	if !strings.Contains(rowA, "presencelist__dot-online") {
		t.Errorf("expected dot-online for Alice, got %s", rowA)
	}
	if !strings.Contains(rowA, "En línea") {
		t.Errorf("expected status text En línea for Alice, got %s", rowA)
	}

	if !strings.Contains(rowB, "presencelist__dot-offline") {
		t.Errorf("expected dot-offline for Bob, got %s", rowB)
	}
	if !strings.Contains(rowB, "Desconectado") {
		t.Errorf("expected status text Desconectado for Bob, got %s", rowB)
	}
}
