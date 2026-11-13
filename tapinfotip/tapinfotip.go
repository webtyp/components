// Package tapinfotip is the help control of a form field: a small (i) that
// shows its help text only when it is tapped, and hides it again on the next
// tap. Nothing is drawn for the text until someone asks for it.
//
// A free-standing use calls Render. A form places the button in the label row
// of a field and the panel under the field's control, so it uses Button and
// Panel, which share one id: the button's aria-controls names the panel.
package tapinfotip

import (
	. "webtyp.com/dom"
	"webtyp.com/icons/circleinfo"
	"webtyp.com/lang"
	"webtyp.com/widget"
)

// NameTapInfotip is the widget name for tapinfotip.
const NameTapInfotip = widget.Name("tapinfotip")

const (
	PartButton = widget.Part("button") // the (i) control
	PartIcon   = widget.Part("icon")   // the glyph inside the button
	PartPanel  = widget.Part("panel")  // the help text, shown while open
)

var (
	clsRoot   = NameTapInfotip.Root()
	clsButton = NameTapInfotip.Class(PartButton)
	clsIcon   = NameTapInfotip.Class(PartIcon)
	clsPanel  = NameTapInfotip.Class(PartPanel)
)

// textMoreInfo is the accessible name of the (i): it has no visible words.
const textMoreInfo lang.Text = "More information"

// ARIA values of aria-expanded.
const (
	attrTrue  = "true"
	attrFalse = "false"
)

// TapInfotip is a help (i) with its text. Help is an English key, translated
// when drawn. The text is hidden until the button is tapped; a second tap hides
// it again.
//
// Example:
//
//	tip := &tapinfotip.TapInfotip{Help: "Document number, as written on the document."}
//	label.Child(tip.Button())
//	field.Child(tip.Panel())
type TapInfotip struct {
	Element // value embed — NEVER pointer (TinyGo heap constraint)

	Help lang.Text

	open    *SignalBool
	panelID string
}

func (t *TapInfotip) WidgetName() widget.Name { return NameTapInfotip }
func (t *TapInfotip) WidgetKind() widget.Kind { return widget.Disclosure }

// Init creates the open state and the id the button and the panel share.
// Calling it again is a no-op. Button and Panel call it too, so a caller that
// never calls Init still gets a working control.
func (t *TapInfotip) Init(_ Ctx) {
	if t.open != nil {
		return
	}
	t.open = NewBool(false)
	t.panelID = NewElement("div").GetID()
}

// Button is the (i) control. A tap toggles the panel, and aria-expanded follows
// it. The glyph is the shared information icon, drawn in currentColor.
func (t *TapInfotip) Button() *Element {
	t.Init(nil)
	return NewElement("button").Set(clsButton.AsAttr()).
		Attr("type", "button").
		Attr("aria-label", lang.Translate(textMoreInfo).String()).
		Attr("aria-controls", t.panelID).
		BindAttrFunc("aria-expanded", t.expanded).
		Child(circleinfo.Ref.Render(string(clsIcon))).
		OnClick(func(Event) { t.toggle() })
}

// Panel is the help text. It is hidden until Button opens it.
func (t *TapInfotip) Panel() *Element {
	t.Init(nil)
	return NewElement("p").ID(t.panelID).Set(clsPanel.AsAttr()).
		BindState(widget.Open, t.open).
		Text(lang.Translate(t.Help).String())
}

// Render is the free-standing form: the button and the panel in one wrapper.
func (t *TapInfotip) Render() *Element {
	return NewElement("div").Set(clsRoot.AsAttr()).
		Child(t.Button(), t.Panel())
}

func (t *TapInfotip) expanded() string {
	if t.open.Get() {
		return attrTrue
	}
	return attrFalse
}

// toggle flips the panel. When it opens, the panel is brought into view
// (block: nearest, from dom), so a help text under the bottom field is not
// cut off by the screen edge.
func (t *TapInfotip) toggle() {
	t.open.Toggle()
	if !t.open.Get() {
		return
	}
	if ref, ok := Get(t.panelID); ok && ref != nil {
		ref.ScrollIntoView()
	}
}
