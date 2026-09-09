//go:build !wasm

package actionbutton

import (
	"webtyp.com/css"
	"webtyp.com/widget/style"
)

// RenderCSS defines the actionbutton visual contract using the style DSL.
func (b *ActionButton) RenderCSS() *css.Stylesheet {
	return style.For(b).
		// The root is the pressed element itself — the markup is
		// class="actionbutton actionbutton__primary", both rules on one node —
		// so it may not be ruleless (TestPairMarkupAndStylesheet). What belongs
		// here is the one thing the variants cannot supply: with Href set this
		// renders an <a>, which unlike a native <button> does not centre its own
		// label inside the padding Button() adds.
		Root(
			style.CenterContent(),
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
