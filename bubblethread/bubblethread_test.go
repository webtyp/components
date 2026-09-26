//go:build !wasm

package bubblethread

import (
	"strings"
	"testing"
)

func TestBubbleThread_IdempotentAndRootClass(t *testing.T) {
	bt := &BubbleThread{}
	s1 := bt.Render().String()
	s2 := bt.Render().String()
	if s1 != s2 {
		t.Fatalf("expected idempotent Render(), got %q vs %q", s1, s2)
	}
	if !strings.Contains(s1, "bubblethread") {
		t.Fatalf("expected output to contain root class bubblethread, got %q", s1)
	}
}

func TestBubbleThread_RenderCSS(t *testing.T) {
	bt := &BubbleThread{}
	sheet := bt.RenderCSS()
	if sheet == nil {
		t.Fatal("RenderCSS returned nil")
	}
}

func TestBubbleThread_BubblesAndEscaping(t *testing.T) {
	bt := &BubbleThread{
		ReadLabel: "Leído",
		Empty:     "Sin mensajes",
	}

	bMineUnread := Bubble{ID: "b1", Body: "<b>mine unread</b>", Time: "10:00", Mine: true, Read: false}
	bMineRead := Bubble{ID: "b2", Body: "mine read", Time: "10:01", Mine: true, Read: true}
	bTheirs := Bubble{ID: "b3", Author: "Alice", Body: "hello", Time: "10:02", Mine: false}

	bt.SetBubbles([]Bubble{bMineUnread, bMineRead, bTheirs})

	if len(bt.Bubbles()) != 3 {
		t.Fatalf("expected 3 bubbles, got %d", len(bt.Bubbles()))
	}

	outMineUnread := bt.buildBubble(bMineUnread).String()
	outMineRead := bt.buildBubble(bMineRead).String()
	outTheirs := bt.buildBubble(bTheirs).String()

	// Mine unread: has mine part, no author, no ReadLabel, escaped HTML
	if !strings.Contains(outMineUnread, "bubblethread__mine") {
		t.Errorf("expected bubblethread__mine in mine bubble: %s", outMineUnread)
	}
	if strings.Contains(outMineUnread, "<b>") {
		t.Errorf("HTML tags in body must be escaped: %s", outMineUnread)
	}
	if !strings.Contains(outMineUnread, "&lt;b&gt;") {
		t.Errorf("expected escaped HTML &lt;b&gt;: %s", outMineUnread)
	}
	if strings.Contains(outMineUnread, "Leído") {
		t.Errorf("unread bubble must not show ReadLabel: %s", outMineUnread)
	}

	// Mine read: has ReadLabel "Leído"
	if !strings.Contains(outMineRead, "Leído") {
		t.Errorf("read mine bubble must show ReadLabel: %s", outMineRead)
	}

	// Theirs: has theirs part, author "Alice"
	if !strings.Contains(outTheirs, "bubblethread__theirs") {
		t.Errorf("expected bubblethread__theirs in theirs bubble: %s", outTheirs)
	}
	if !strings.Contains(outTheirs, "Alice") {
		t.Errorf("expected author Alice in theirs bubble: %s", outTheirs)
	}

	// Append existing ID: should skip duplicate
	bt.Append(Bubble{ID: "b3", Author: "Alice", Body: "hello", Time: "10:02", Mine: false})
	if len(bt.Bubbles()) != 3 {
		t.Errorf("expected 3 bubbles after appending duplicate ID, got %d", len(bt.Bubbles()))
	}

	// Append new ID
	bt.Append(Bubble{ID: "b4", Body: "new msg", Time: "10:03", Mine: true})
	if len(bt.Bubbles()) != 4 {
		t.Errorf("expected 4 bubbles after appending new ID, got %d", len(bt.Bubbles()))
	}

	// MarkRead
	bt.MarkRead("b1")
	if !bt.Bubbles()[0].Read {
		t.Errorf("b1 should be marked as Read after MarkRead")
	}
}
