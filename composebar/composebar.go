package composebar

import (
	. "webtyp.com/dom"
	"webtyp.com/fmt"
	. "webtyp.com/html"
	"webtyp.com/lang"
	"webtyp.com/widget"
)

// NameComposeBar is the widget identity.
const NameComposeBar = widget.Name("composebar")

const (
	PartBar   = widget.Part("bar")
	PartInput = widget.Part("input")
	PartSend  = widget.Part("send")
)

var (
	clsBar   = NameComposeBar.Root()
	clsInput = NameComposeBar.Class(PartInput)
	clsSend  = NameComposeBar.Class(PartSend)
)

type ComposeBar struct {
	Element
	// English; translated in Render.
	Placeholder lang.Text
	// English; translated in Render.
	SendLabel lang.Text         // button text, e.g. "Send"
	MaxLength int               // > 0 sets maxlength on the input; <= 0 sets none
	OnSend    func(body string) // called with the trimmed, non-empty text
	Disabled  *SignalBool       // optional; nil = always enabled

	text *SignalString
}

func (c *ComposeBar) WidgetName() widget.Name { return NameComposeBar }
func (c *ComposeBar) WidgetKind() widget.Kind { return widget.Form }

func (c *ComposeBar) ensure() {
	if c.text == nil {
		c.text = NewString("")
	}
}

func (c *ComposeBar) Init(_ Ctx) { c.ensure() }

func (c *ComposeBar) send() {
	c.ensure()
	body := fmt.TrimSpace(c.text.Get())
	if body == "" {
		return
	}
	if c.OnSend != nil {
		c.OnSend(body)
	}
	c.text.Set("")
}

func (c *ComposeBar) Render() *Element {
	if c.OnSend == nil {
		panic("composebar: OnSend is required")
	}
	c.ensure()

	input := Input("text").Set(clsInput.AsAttr()).
		Bind(c.text)

	if c.Placeholder != "" {
		input.Attr("placeholder", lang.Translate(c.Placeholder).String())
	}
	if c.MaxLength > 0 {
		input.Attr("maxlength", fmt.Sprint(c.MaxLength))
	}
	if c.Disabled != nil {
		input.BindAttrBool("disabled", c.Disabled).
			BindState(widget.Disabled, c.Disabled)
	}

	input.OnKeyDown(func(e KeyEvent) {
		if e.Key() == KeyEnter {
			e.PreventDefault()
			c.send()
		}
	})

	sendBtn := Button().Set(clsSend.AsAttr()).
		Attr("type", "button").
		Text(lang.Translate(c.SendLabel).String())

	if c.Disabled != nil {
		sendBtn.BindAttrBool("disabled", c.Disabled).
			BindState(widget.Disabled, c.Disabled)
	}

	sendBtn.OnClick(func(Event) {
		c.send()
	})

	return NewElement("form").Set(clsBar.AsAttr()).
		Child(input).
		Child(sendBtn)
}
