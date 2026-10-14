//go:build !wasm

package modaldialog

import (
	"webtyp.com/css"
	"webtyp.com/widget/style"
)

// RenderCSS defines the modaldialog visual contract using the style DSL.
func (m *ModalDialog) RenderCSS() *css.Stylesheet {
	return style.For(m).
		// Backdrop(Viewport), not Fill(): in the flow the dialog was a flex item
		// of whatever mounted it and got dealt a column of that layout — a
		// sliver between the form and the list. A modal answers to the screen,
		// not to the box it happens to be mounted in.
		//
		// The wash (Veil) is this element's own background and the panel is its
		// in-flow child: a child always paints above its parent's background,
		// so the panel sits above the wash by construction — no sibling to
		// outrank, no stacking level to declare. Clicking the wash closes the
		// dialog (the panel stops the click from reaching the root; see
		// Render()). There is no click-catcher part anymore: a positioned
		// sibling is the one element the DSL cannot place under an in-flow
		// peer, which is why the catcher had to paint above the panel.
		Root(
			style.Backdrop(style.Viewport),
			style.Veil(),
			style.CenterContent(),
			style.Pad(style.Space4),
		).
		Part(PartPanel,
			style.As(style.Panel),
			style.Round(style.RadiusLg),
			style.Raise(style.Popover),
			style.Pad(style.Space3),
			style.Stack(style.Space2),
			style.Width(style.Content),
		).
		Part(PartHeader,
			style.Row(style.Space2),
		).
		Part(PartTitle,
			style.Heading(),
		).
		Part(PartBody,
			style.Stack(style.Space1),
		).
		Part(PartClose,
			style.As(style.Panel),
		).
		Stylesheet()
}
