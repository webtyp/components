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

func TestTwoInstancesShareAPage(t *testing.T) {
	sn1 := &sitenav.SiteNav{
		Links: []sitenav.NavItem{{Label: "A", Href: "/a"}},
	}
	sn2 := &sitenav.SiteNav{
		Links: []sitenav.NavItem{{Label: "B", Href: "/b"}},
	}
	sn1.Init(nil)
	sn2.Init(nil)

	parent := NewElement("div").Child(sn1).Child(sn2)
	if err := Render("app", parent); err != nil {
		t.Fatalf("mounting two sitenavs in one render failed: %v", err)
	}

	navs := js.Global().Get("document").Call("querySelectorAll", ".sitenav")
	if navs.Get("length").Int() != 2 {
		t.Fatalf("expected 2 sitenavs in DOM, got %d", navs.Get("length").Int())
	}
}
