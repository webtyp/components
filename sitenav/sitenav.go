package sitenav

import (
	. "webtyp.com/dom"
	. "webtyp.com/html"
	"webtyp.com/image"
	"webtyp.com/svg"
	"webtyp.com/widget"
)

// NameSiteNav is the widget name for sitenav.
const NameSiteNav = widget.Name("sitenav")

const (
	PartBrand     = widget.Part("brand")
	PartLogo      = widget.Part("logo")
	PartToggle    = widget.Part("toggle")
	PartMenu      = widget.Part("menu")
	PartNav       = widget.Part("nav")
	PartLink      = widget.Part("link")
	PartActions   = widget.Part("actions")
	PartIconOpen  = widget.Part("icon-open")
	PartIconClose = widget.Part("icon-close")
)

var (
	clsNav          = NameSiteNav.Root()
	clsNavBrand     = NameSiteNav.Class(PartBrand)
	clsNavLogo      = NameSiteNav.Class(PartLogo)
	clsNavToggle    = NameSiteNav.Class(PartToggle)
	clsNavMenu      = NameSiteNav.Class(PartMenu)
	clsNavNav       = NameSiteNav.Class(PartNav)
	clsNavLink      = NameSiteNav.Class(PartLink)
	clsNavActions   = NameSiteNav.Class(PartActions)
	clsNavIconOpen  = NameSiteNav.Class(PartIconOpen)
	clsNavIconClose = NameSiteNav.Class(PartIconClose)
)

// menuID is the element the toggle controls. Sprite symbol IDs and element
// IDs share one document-wide namespace, and the hamburger glyph used to be
// called "sitenav-menu" too: the sprite is injected at the top of <body>, so
// getElementById("sitenav-menu") returned the <symbol>, and the toggle wrote
// the open state onto an invisible SVG node while the real menu never moved.
const menuID = "sitenav-menu"

const (
	iconMenu  = svg.Icon("sitenav-hamburger")
	iconClose = svg.Icon("sitenav-close")
)

// NavItem represents a single link item in the site navigation bar.
type NavItem struct {
	Label  string
	Href   string
	Active bool
}

// SiteNav provides the main site header with brand logo, navigation links,
// actions, and a responsive mobile toggle menu.
type SiteNav struct {
	Element
	WideLogoSrc    string
	CompactLogoSrc string
	LogoAlt        string
	Links          []NavItem
	Actions        []Component
}

func (sn *SiteNav) WidgetName() widget.Name { return NameSiteNav }

// WidgetKind is Disclosure, not Region: the markup this component already
// emits IS the disclosure pattern — a button carrying aria-expanded and
// aria-controls that shows and hides the region it names. Region has no
// expandable state at all (widget.Kind.Allows), so declaring it left the
// collapsed mobile menu inexpressible: the toggle rendered and did nothing a
// stylesheet could react to.
func (sn *SiteNav) WidgetKind() widget.Kind { return widget.Disclosure }

func (sn *SiteNav) Render() *Element {
	header := Header().Set(clsNav.AsAttr())

	// Brand section with responsive <picture> logo
	brand := Div().Set(clsNavBrand.AsAttr())
	pic := image.Picture()
	if sn.WideLogoSrc != "" {
		pic.Child(image.Source(sn.WideLogoSrc, "(min-width: 768px)"))
	}
	src := sn.CompactLogoSrc
	if src == "" {
		src = sn.WideLogoSrc
	}
	pic.Child(image.Img(src, sn.LogoAlt).Class(string(clsNavLogo)).AsElement())
	brand.Child(pic)
	header.Child(brand)

	// Toggle button for mobile navigation
	toggleBtn := Button().Set(clsNavToggle.AsAttr()).
		Attr("aria-controls", menuID).
		Attr("aria-expanded", "false").
		Attr("aria-label", "Abrir menú de navegación").
		Child(iconMenu.Render(clsNavIconOpen.String())).
		Child(iconClose.Render(clsNavIconClose.String()))
	header.Child(toggleBtn)

	// Menu container.
	//
	// ID(), not Key()+Ref(): this component's behaviour ships as JavaScript
	// (see RenderJS in js.go), which reaches the menu with
	// document.getElementById("sitenav-menu") in three places and matches the
	// toggle with [aria-controls="sitenav-menu"]. A dom-assigned id is invisible
	// to that script, so replacing this with a Key silently breaks the mobile
	// menu — it simply never opens, with no error.
	//
	// A site nav is a page singleton by construction (the script itself guards
	// on window.__sitenavInit), so a fixed global id is the case ID() exists
	// for — the same reason platformd keeps pd-msg-slot and pd-hamburger-btn.
	menu := Div().Set(clsNavMenu.AsAttr()).ID(menuID)

	// PartNav is declared in RenderCSS; without the class here every rule it
	// carries addressed nothing, and the links kept the menu's own layout.
	nav := Nav().Set(clsNavNav.AsAttr())
	for _, link := range sn.Links {
		a := A(link.Href).Set(clsNavLink.AsAttr()).Text(link.Label)
		if link.Active {
			a.Attr("aria-current", "page")
		}
		nav.Child(a)
	}
	menu.Child(nav)

	if len(sn.Actions) > 0 {
		actions := Div().Set(clsNavActions.AsAttr())
		for _, act := range sn.Actions {
			if act != nil {
				actions.Child(act)
			}
		}
		menu.Child(actions)
	}

	header.Child(menu)
	return header
}
