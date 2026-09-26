//go:build !wasm

package inboxlist

import (
	"strings"
	"testing"
)

func TestInboxList_IdempotentAndRootClass(t *testing.T) {
	l := &InboxList{}
	s1 := l.Render().String()
	s2 := l.Render().String()
	if s1 != s2 {
		t.Fatalf("expected idempotent Render(), got %q vs %q", s1, s2)
	}
	if !strings.Contains(s1, "inboxlist") {
		t.Fatalf("expected output to contain root class inboxlist, got %q", s1)
	}
}

func TestInboxList_RenderCSS(t *testing.T) {
	l := &InboxList{}
	sheet := l.RenderCSS()
	if sheet == nil {
		t.Fatal("RenderCSS returned nil")
	}
}

func TestInboxList_SetRows(t *testing.T) {
	l := &InboxList{
		Empty: "No messages",
	}

	rows := []Row{
		{ID: "r1", Title: "Read item", Preview: "hello", Time: "10:00", Unread: 0},
		{ID: "r2", Title: "Unread item", Preview: "world", Time: "10:05", Unread: 3},
	}
	l.SetRows(rows)

	if len(l.Rows()) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(l.Rows()))
	}

	row1HTML := l.buildRow(rows[0]).String()
	row2HTML := l.buildRow(rows[1]).String()

	// Check row classes and title variants
	if !strings.Contains(row1HTML, "inboxlist__title") {
		t.Errorf("expected inboxlist__title part on read row, got %q", row1HTML)
	}
	if strings.Contains(row1HTML, "inboxlist__title-unread") {
		t.Errorf("did not expect inboxlist__title-unread part on read row, got %q", row1HTML)
	}

	if !strings.Contains(row2HTML, "inboxlist__title-unread") {
		t.Errorf("expected inboxlist__title-unread part on unread row, got %q", row2HTML)
	}

	// Unread = 3 row: countbadge should have count 3
	if !strings.Contains(row2HTML, "3") {
		t.Errorf("expected badge text '3' in unread row output, got %q", row2HTML)
	}
}
