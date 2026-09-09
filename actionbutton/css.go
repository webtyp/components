//go:build !wasm

package actionbutton

import (
	"webtyp.com/css"
	"webtyp.com/widget/style"
)

// RenderCSS defines the actionbutton visual contract using the style DSL.
func (b *ActionButton) RenderCSS() *css.Stylesheet {
	return style.For(b).
		Root(
			style.KeepSize(),
		).
		Part(PartPrimary,
			style.Button(style.Primary),
		).
		Part(PartSecondary,
			style.Button(style.Secondary),
		).
		Part(PartDanger,
			style.Button(style.Danger),
		).
		Stylesheet()
}
