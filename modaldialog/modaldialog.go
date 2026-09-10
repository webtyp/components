package modaldialog

import (
	. "webtyp.com/dom"
	. "webtyp.com/html"
	"webtyp.com/widget"
)

// NameModalDialog is the widget name for modal dialog.
const NameModalDialog = widget.Name("modaldialog")

const (
	PartPanel   = widget.Part("panel")
	PartHeader  = widget.Part("header")
	PartBody    = widget.Part("body")
	PartActions = widget.Part("actions")
	PartClose   = widget.Part("close")
)

var (
	clsModal        = NameModalDialog.Root()
	clsModalContent = NameModalDialog.Class(PartPanel)
	clsModalHeader  = NameModalDialog.Class(PartHeader)
	clsModalBody    = NameModalDialog.Class(PartBody)
	clsModalClose   = NameModalDialog.Class(PartClose)
)

// ModalDialog represents a modal dialog component.
type ModalDialog struct {
	Element
	Title   string
	Content Component

	// HideClose drops the "×" from the header. Set it when the dialog's own
	// content already offers an explicit way out — a confirmation with
	// Cancel/Confirm buttons, where a third exit is one more thing to read and
	// says nothing Cancel does not. Clicking outside the panel still closes.
	HideClose bool

	visible *SignalBool
}

func (m *ModalDialog) WidgetName() widget.Name { return NameModalDialog }
func (m *ModalDialog) WidgetKind() widget.Kind { return widget.Dialog }

func (m *ModalDialog) Init(_ Ctx) {
	m.visible = NewBool(false)
}

func (m *ModalDialog) Render() *Element {
	header := Div().Set(clsModalHeader.AsAttr()).
		Child(H2().Text(m.Title))
	if !m.HideClose {
		header.Child(Button().
			Text("×").
			Set(clsModalClose.AsAttr()).
			OnClick(func(e Event) { m.visible.Set(false) }))
	}

	modalContent := Div().Set(clsModalContent.AsAttr()).
		Child(header).
		Child(
			Div().Set(clsModalBody.AsAttr()).Child(m.Content),
		)

	// The root IS the wash, and it is also the click target: a click that
	// reaches it landed outside the panel, so it closes the dialog. The panel
	// stops the click first — its whole subtree, close button included — so
	// interacting with the dialog never closes it. There is deliberately no
	// separate click-catcher element: a positioned sibling would have to
	// paint BELOW the panel, and the DSL has no level that puts an in-flow
	// element above a positioned one. The panel instead wins by being the
	// wash's own in-flow child — a child always paints above its parent's
	// background — so the click model needs no stacking at all.
	modal := Div().Set(clsModal.AsAttr()).
		Attr("role", "dialog").
		Attr("aria-modal", "true").
		OnClick(func(e Event) { m.visible.Set(false) }).
		Child(modalContent)

	modalContent.OnClick(func(e Event) { e.StopPropagation() })
	return Show(m.visible, modal)
}

func (m *ModalDialog) Open() {
	m.visible.Set(true)
}

// Close hides the dialog programmatically.
func (m *ModalDialog) Close() {
	m.visible.Set(false)
}
