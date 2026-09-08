//go:build wasm

package sitenav_test

import (
	"syscall/js"
	"testing"

	"webtyp.com/components/sitenav"
	. "webtyp.com/dom"
)

func TestMain(m *testing.M) {
	app := js.Global().Get("document").Call("createElement", "div")
	app.Set("id", "app")
	js.Global().Get("document").Get("body").Call("appendChild", app)
	m.Run()
}

// TestMenuKeepsTheIDItsScriptLooksUp is the regression for this component's
// real contract: SiteNav's behaviour ships as JavaScript (RenderJS in js.go),
// which finds the menu with document.getElementById("sitenav-menu") and the
// toggle with [aria-controls="sitenav-menu"]. Swapping that id for a
// dom-assigned one breaks the mobile menu SILENTLY — the script's `if (menu)`
// guards simply stop matching and nothing ever opens.
//
// A site nav is a page singleton (the script guards on window.__sitenavInit),
// so the fixed global id is correct here. This test pins it.
func TestMenuKeepsTheIDItsScriptLooksUp(t *testing.T) {
	sn := &sitenav.SiteNav{
		Links: []sitenav.NavItem{{Label: "A", Href: "/a"}},
	}

	if err := Render("app", sn); err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	doc := js.Global().Get("document")

	menu := doc.Call("getElementById", "sitenav-menu")
	if menu.IsNull() || menu.IsUndefined() {
		t.Fatal(`getElementById("sitenav-menu") found nothing — js.go reaches the menu this way in three places; the mobile menu cannot open`)
	}

	toggle := doc.Call("querySelector", `[aria-controls="sitenav-menu"]`)
	if toggle.IsNull() || toggle.IsUndefined() {
		t.Fatal(`no element carries aria-controls="sitenav-menu" — js.go matches the toggle with this selector`)
	}

	// The relation must actually resolve: aria-controls has to name a node that
	// exists, or the attribute is a dangling reference to assistive tech too.
	if got := toggle.Call("getAttribute", "aria-controls").String(); got != menu.Get("id").String() {
		t.Errorf("aria-controls=%q does not match the menu's id %q", got, menu.Get("id").String())
	}
}
