//go:build wasm

package composebar_test

import (
	"syscall/js"
	"testing"

	"webtyp.com/components/composebar"
	. "webtyp.com/dom"
)

func TestMain(m *testing.M) {
	app := js.Global().Get("document").Call("createElement", "div")
	app.Set("id", "app")
	js.Global().Get("document").Get("body").Call("appendChild", app)
	m.Run()
}

func TestComposeBar_SendWASM(t *testing.T) {
	var sent []string
	cb := &composebar.ComposeBar{
		Placeholder: "Escribe...",
		SendLabel:   "Enviar",
		OnSend: func(body string) {
			sent = append(sent, body)
		},
	}

	if err := Render("app", cb); err != nil {
		t.Fatalf("mounting composebar failed: %v", err)
	}

	doc := js.Global().Get("document")
	input := doc.Call("querySelector", "input[type='text']")
	sendBtn := doc.Call("querySelector", "button[type='button']")

	if input.IsNull() || input.IsUndefined() {
		t.Fatal("input element not found in DOM")
	}
	if sendBtn.IsNull() || sendBtn.IsUndefined() {
		t.Fatal("send button not found in DOM")
	}

	// 1. Send empty text via button click -> should do nothing
	sendBtn.Call("click")
	if len(sent) != 0 {
		t.Fatalf("expected 0 sent messages on empty click, got %d", len(sent))
	}

	// 2. Set input value to "  hola " and trigger input event
	input.Set("value", "  hola ")
	input.Call("dispatchEvent", js.Global().Get("Event").New("input"))

	// Click send button
	sendBtn.Call("click")

	if len(sent) != 1 || sent[0] != "hola" {
		t.Fatalf("expected sent ['hola'], got %v", sent)
	}

	// Check input cleared
	if val := input.Get("value").String(); val != "" {
		t.Errorf("expected input to be cleared after send, got %q", val)
	}
}
