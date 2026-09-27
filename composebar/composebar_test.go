//go:build !wasm

package composebar

import (
	"strings"
	"testing"
)

func TestComposeBar_IdempotentAndRootClass(t *testing.T) {
	c := &ComposeBar{
		OnSend: func(string) {},
	}
	s1 := c.Render().String()
	s2 := c.Render().String()
	if s1 != s2 {
		t.Fatalf("expected idempotent Render(), got %q vs %q", s1, s2)
	}
	if !strings.Contains(s1, "composebar") {
		t.Fatalf("expected output to contain root class composebar, got %q", s1)
	}
}

func TestComposeBar_RenderCSS(t *testing.T) {
	c := &ComposeBar{
		OnSend: func(string) {},
	}
	sheet := c.RenderCSS()
	if sheet == nil {
		t.Fatal("RenderCSS returned nil")
	}
}

func TestComposeBar_NilOnSendPanics(t *testing.T) {
	c := &ComposeBar{}
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic when OnSend is nil")
		}
		if msg, ok := r.(string); !ok || msg != "composebar: OnSend is required" {
			t.Fatalf("unexpected panic message: %v", r)
		}
	}()
	c.Render()
}

func TestComposeBar_MaxLengthAndPlaceholder(t *testing.T) {
	c := &ComposeBar{
		Placeholder: "Type a message...",
		SendLabel:   "Send",
		MaxLength:   2000,
		OnSend:      func(string) {},
	}

	html := c.Render().String()

	if !strings.Contains(html, "maxlength='2000'") && !strings.Contains(html, "maxlength=\"2000\"") {
		t.Errorf("expected maxlength 2000 in output, got: %s", html)
	}
	if !strings.Contains(html, "placeholder='Type a message...'") && !strings.Contains(html, "placeholder=\"Type a message...\"") {
		t.Errorf("expected placeholder in output, got: %s", html)
	}
}
