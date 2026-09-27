//go:build wasm

package bubblethread_test

import (
	"syscall/js"
	"testing"

	"webtyp.com/components/bubblethread"
	. "webtyp.com/dom"
)

func TestMain(m *testing.M) {
	app := js.Global().Get("document").Call("createElement", "div")
	app.Set("id", "app")
	js.Global().Get("document").Get("body").Call("appendChild", app)
	m.Run()
}

func TestBubbleThread_WASM(t *testing.T) {
	bt := &bubblethread.BubbleThread{
		ReadLabel: "Leído",
	}

	bt.SetBubbles([]bubblethread.Bubble{
		{ID: "m1", Body: "Hello", Time: "10:00", Mine: false, Author: "Alice"},
	})

	if err := Render("app", bt); err != nil {
		t.Fatalf("mounting bubblethread failed: %v", err)
	}

	bt.Append(bubblethread.Bubble{ID: "m2", Body: "Hi Alice", Time: "10:01", Mine: true})

	doc := js.Global().Get("document")
	lastBubble := doc.Call("querySelector", "[data-id='m2']")
	if lastBubble.IsNull() || lastBubble.IsUndefined() {
		t.Fatal("expected last bubble element with data-id m2 in DOM")
	}
}
