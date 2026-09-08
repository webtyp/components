//go:build !wasm

package sitenav

import (
	"strings"
	"testing"

	"webtyp.com/js"
	"webtyp.com/widget"
)

func TestSiteNav_AccessibilityAttributes(t *testing.T) {
	sn := &SiteNav{
		WideLogoSrc:    "/logo.png",
		CompactLogoSrc: "/logo-sm.png",
		LogoAlt:        "Logo",
		Links: []NavItem{
			{Label: "Home", Href: "/"},
		},
	}

	html := sn.Render().String()

	if !strings.Contains(html, `aria-controls='sitenav-menu'`) {
		t.Errorf("expected aria-controls='sitenav-menu' in markup, got: %s", html)
	}

	if !strings.Contains(html, `aria-expanded='false'`) {
		t.Errorf("expected aria-expanded='false' in markup, got: %s", html)
	}

}

func TestSiteNav_RenderJSIdempotentAndSafe(t *testing.T) {
	sn := &SiteNav{}
	scripts := sn.RenderJS()

	if len(scripts) != 1 || scripts[0].Content == "" {
		t.Fatal("RenderJS returned no script content")
	}
	jsStr := scripts[0].Content

	if !strings.Contains(jsStr, "__sitenavInit") {
		t.Errorf("RenderJS does not include idempotency guard __sitenavInit")
	}

	if !strings.Contains(jsStr, "document.addEventListener") {
		t.Errorf("RenderJS does not use document event delegation")
	}
}

func TestSiteNav_RenderJSSatisfiesSSRJSProvider(t *testing.T) {
	// sitec.RegisterComponents type-asserts for RenderJS() []*js.Script at
	// runtime; a value-slice return silently fails that assertion instead of
	// erroring, so this pins the pointer-slice shape as a compile-time check.
	var _ interface {
		RenderJS() []*js.Script
	} = (*SiteNav)(nil)
}

// TestSiteNav_MobileMenuCollapses pins the three halves of the responsive nav
// that have to agree for it to work at all: the hamburger exists only where
// the links do not fit, the links are hidden on that same viewport, and the
// open state is the one the toggle actually writes. Any one of them drifting
// alone is silent — the button renders, the click handler runs, and nothing
// on screen changes.
func TestSiteNav_MobileMenuCollapses(t *testing.T) {
	sn := &SiteNav{Links: []NavItem{{Label: "Inicio", Href: "#inicio"}}}

	sheet := sn.RenderCSS().String()

	if !strings.Contains(sheet, "."+clsNavMenu.String()+"["+widget.Open.Key()+`="`+widget.Open.Value()+`"]`) {
		t.Errorf("menu is never revealed by the Open state the toggle writes:\n%s", sheet)
	}

	script := sn.RenderJS()[0].Content
	if !strings.Contains(script, widget.Open.Key()) {
		t.Errorf("toggle JS does not write %s, the attribute RenderCSS selects on:\n%s", widget.Open.Key(), script)
	}
	if strings.Contains(script, "is-open") {
		t.Errorf("toggle JS still flips an is-open class that no stylesheet matches:\n%s", script)
	}
}

// TestSiteNav_IconIDsDoNotCollideWithElementIDs guards the namespace the
// sprite shares with the document. A <symbol id="x"> and an element id="x" are
// the same id as far as getElementById is concerned, and the sprite is
// injected first — so a collision silently hands every lookup to an invisible
// SVG node.
func TestSiteNav_IconIDsDoNotCollideWithElementIDs(t *testing.T) {
	sn := &SiteNav{}
	for _, def := range sn.IconSvg().Icons() {
		if def.Icon.ID() == menuID {
			t.Errorf("icon %q shares its id with the menu element; getElementById(%q) will return the sprite symbol instead of the menu", def.Icon.ID(), menuID)
		}
	}
}

// TestSiteNav_ToggleGlyphsAreSizedAndSwapped guards the control's two failure
// modes, both of which shipped: an <svg> with no box falls back to the
// replaced-element default of 300x150 and spills out of the button, and with
// nothing choosing between them the hamburger and the cross paint at once.
func TestSiteNav_ToggleGlyphsAreSizedAndSwapped(t *testing.T) {
	sn := &SiteNav{Links: []NavItem{{Label: "Inicio", Href: "#inicio"}}}

	sheet := sn.RenderCSS().String()

	for _, cls := range []string{clsNavIconOpen.String(), clsNavIconClose.String()} {
		if !strings.Contains(sheet, "."+cls+" {") {
			t.Errorf("glyph %q has no rule at all, so it renders at the 300x150 replaced-element default:\n%s", cls, sheet)
		}
	}

	openState := "." + clsNavToggle.String() + "[" + widget.Open.Key() + `="` + widget.Open.Value() + `"] .` + clsNavIconClose.String()
	if !strings.Contains(sheet, openState) {
		t.Errorf("the close glyph is never revealed while the toggle is open; the button keeps showing the hamburger:\n%s", sheet)
	}

	script := sn.RenderJS()[0].Content
	if !strings.Contains(script, "toggle.setAttribute('"+widget.Open.Key()+"'") {
		t.Errorf("the toggle never carries %s, so the descendant rule above matches nothing:\n%s", widget.Open.Key(), script)
	}
}
