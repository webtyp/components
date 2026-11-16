//go:build wasm

package searchbar

import (
	"syscall/js"
	"testing"

	"webtyp.com/dom"
)

func TestSearchBar_SearchesOnEnterOnly(t *testing.T) {
	doc := js.Global().Get("document")
	mount := doc.Call("createElement", "div")
	mount.Set("id", "sb-mount")
	doc.Get("body").Call("appendChild", mount)

	s := &SearchBar{}
	var terms []string
	typed := 0
	s.OnFilterChange(func(term string) { terms = append(terms, term) })
	s.OnFilterTyping(func() { typed++ })
	if err := dom.Render("sb-mount", s); err != nil {
		t.Fatal(err)
	}
	in := mount.Call("querySelector", "input")
	in.Set("value", "lucia")
	in.Call("dispatchEvent", js.Global().Get("Event").New("input"))
	in.Set("value", "")
	in.Call("dispatchEvent", js.Global().Get("Event").New("input"))
	if len(terms) != 0 || typed != 2 {
		t.Fatalf("typing/clearing must not search: terms=%v typed=%d", terms, typed)
	}
	in.Set("value", "lucia")
	opts := js.Global().Get("Object").New()
	opts.Set("key", "Enter")
	in.Call("dispatchEvent", js.Global().Get("KeyboardEvent").New("keydown", opts))
	if len(terms) != 1 || terms[0] != "lucia" {
		t.Fatalf("Enter must search with the typed term, got %v", terms)
	}
}
