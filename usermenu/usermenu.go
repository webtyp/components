// Package usermenu is the identity surface of an application shell: who is
// logged in, at rest, and what they can do about it, on demand.
package usermenu

import (
	. "webtyp.com/dom"
	"webtyp.com/fmt"
	. "webtyp.com/html"
	"webtyp.com/svg"
	"webtyp.com/widget"
)

// NameUserMenu is the widget identity.
const NameUserMenu = widget.Name("usermenu")

const (
	PartTrigger  = widget.Part("trigger")
	PartAvatar   = widget.Part("avatar")
	PartName     = widget.Part("name")
	PartPanel    = widget.Part("panel")
	PartRoles    = widget.Part("roles")
	PartRole     = widget.Part("role")
	PartActions  = widget.Part("actions")
	PartBackdrop = widget.Part("backdrop")
)

var (
	clsRoot     = NameUserMenu.Root()
	clsTrigger  = NameUserMenu.Class(PartTrigger)
	clsAvatar   = NameUserMenu.Class(PartAvatar)
	clsName     = NameUserMenu.Class(PartName)
	clsPanel    = NameUserMenu.Class(PartPanel)
	clsRoles    = NameUserMenu.Class(PartRoles)
	clsRole     = NameUserMenu.Class(PartRole)
	clsActions  = NameUserMenu.Class(PartActions)
	clsBackdrop = NameUserMenu.Class(PartBackdrop)
)

// roleChars is the chip's budget, calibrated against the --chip-width the skin
// gives it. Truncate counts the three-byte ellipsis inside this number.
//
// The count is BYTES, not runes, so an accented character costs two and a cut
// can land mid-character. Role names are short labels in practice.
const roleChars = 14

// menuGroup makes each UserMenu on a page part of one native "exclusive
// accordion": opening one closes any other. A shell renders two — one for the
// header and one for the drawer — and without this, opening either would leave
// the other open behind a viewport change.
const menuGroup = "usermenu-group"

// UserMenu shows who is logged in and, on demand, the roles they hold and the
// preferences they can change.
//
// It takes plain data, never an interface owned by a shell: this package sits
// below any layout in the dependency graph, and typing the shell's contract
// here would invert that. A shell adapts whatever it knows into these fields.
type UserMenu struct {
	Element

	// Name is who is logged in. The only thing shown at rest beside the avatar.
	Name string

	// Avatar is the URL of their picture. Empty is normal and expected — most
	// systems have no image for most users — and falls back to Fallback.
	Avatar string

	// Roles are display names, not codes. They appear only inside the panel: a
	// shell cannot know how many roles an application defines nor how long
	// their names are, so nothing unbounded may sit in the resting state.
	Roles []string

	// Fallback is the glyph drawn when Avatar is empty. The shell owns it —
	// picking a sprite is a rendering decision and this component has no
	// business making it for its host.
	Fallback svg.Icon

	// Actions holds the user's preferences — a theme toggle, a language picker,
	// a sign-out. Optional; the panel simply omits the section.
	Actions Component

	open *SignalBool
}

func (m *UserMenu) ensure() {
	if m.open == nil {
		m.open = NewBool(false)
	}
}

func (m *UserMenu) Init(_ Ctx) { m.ensure() }

func (m *UserMenu) WidgetName() widget.Name { return NameUserMenu }
func (m *UserMenu) WidgetKind() widget.Kind { return widget.Menu }

func (m *UserMenu) Render() *Element {
	m.ensure()

	trigger := Summary().Set(clsTrigger.AsAttr()).
		Attr("aria-label", m.Name)

	if m.Avatar != "" {
		trigger.Child(NewElement("img").
			Set(clsAvatar.AsAttr()).
			Attr("src", m.Avatar).
			Attr("alt", "").
			Attr("loading", "lazy"))
	} else {
		trigger.Child(m.Fallback.Render(string(clsAvatar)))
	}
	trigger.Child(Span().Set(clsName.AsAttr()).Text(m.Name))

	panel := Div().Set(clsPanel.AsAttr())

	if len(m.Roles) > 0 {
		roles := Div().Set(clsRoles.AsAttr())
		for _, r := range m.Roles {
			roles.Child(Span().Set(clsRole.AsAttr()).
				Attr("title", r). // the untruncated name stays reachable
				Text(fmt.Convert(r).Truncate(roleChars).String()))
		}
		panel.Child(roles)
	}

	if m.Actions != nil {
		panel.Child(Div().Set(clsActions.AsAttr()).Child(m.Actions))
	}

	// <details> writes its native `open` onto ITSELF, so the backdrop cannot
	// select on it — the state has to be mirrored into a signal the sheet can
	// reach. Same mechanism targetlist uses for its row menus.
	backdrop := Div().Set(clsBackdrop.AsAttr()).
		BindStateFunc(widget.Open, func() bool { return m.open.Get() })

	// Order matters and is not cosmetic. The backdrop and the panel both sit on
	// the widget's stacking level, so among equals the LATER one paints on top:
	// with the backdrop last it covered the panel and swallowed every click
	// meant for the controls inside. Backdrop before panel — the order
	// targetlist already uses. The <summary> stays first because that is what
	// makes it the disclosure control.
	menu := Details().Set(clsRoot.AsAttr()).
		Attr("name", menuGroup).
		Child(trigger).
		Child(backdrop).
		Child(panel)

	menu.On("toggle", func(Event) {
		if ref, ok := menu.Ref(); ok {
			m.open.Set(ref.GetAttr("open") != "<null>")
		}
	})

	backdrop.On("click", func(Event) {
		if ref, ok := menu.Ref(); ok {
			ref.RemoveAttr("open")
		}
		m.open.Set(false)
	})

	return menu
}
