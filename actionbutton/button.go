package actionbutton

import (
	. "webtyp.com/dom"
	. "webtyp.com/html"
	"webtyp.com/widget"
)

// NameActionButton is the widget name for actionbutton.
const NameActionButton = widget.Name("actionbutton")

const (
	PartPrimary     = widget.Part("primary")
	PartSecondary   = widget.Part("secondary")
	PartDanger      = widget.Part("danger")
	PartUrlUp       = widget.Part("url-up")
	PartUrlDown     = widget.Part("url-down")
	PartUrl         = widget.Part("url")
	PartUrlDisable  = widget.Part("url-disable")
	PartSelected    = widget.Part("selected")
	PartLogin       = widget.Part("login")
	PartUrlPulse    = widget.Part("url-pulse")
	PartContebuton  = widget.Part("contebuton")
	PartCenteredBtn = widget.Part("centered-btn")
)

var (
	clsBtn          = NameActionButton.Root()
	clsBtnPrimary   = NameActionButton.Class(PartPrimary)
	clsBtnSecondary = NameActionButton.Class(PartSecondary)
	clsBtnDanger    = NameActionButton.Class(PartDanger)
)

type ActionButton struct {
	Element
	Text    string
	Variant string // "primary", "secondary", "danger"

	// Href renders the button as <a href> instead of <button>: no click
	// handler, works before WASM loads and with JavaScript disabled. Set
	// this for navigation (e.g. an OAuth login link); OnClick is ignored
	// when Href is non-empty.
	Href string

	OnClick func(Event)
}

func (b *ActionButton) WidgetName() widget.Name { return NameActionButton }
func (b *ActionButton) WidgetKind() widget.Kind { return widget.Region }

func (b *ActionButton) Render() *Element {
	var variantCls widget.Class
	switch b.Variant {
	case "secondary":
		variantCls = clsBtnSecondary
	case "danger":
		variantCls = clsBtnDanger
	default:
		variantCls = clsBtnPrimary
	}

	cls := string(clsBtn) + " " + string(variantCls)

	if b.Href != "" {
		return A(b.Href).Attr("class", cls).Text(b.Text)
	}

	btn := Button().
		Attr("class", cls).
		Text(b.Text)

	if b.OnClick != nil {
		btn.OnClick(b.OnClick)
	}

	return btn
}
