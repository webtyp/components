//go:build !wasm

package chipgrid

import (
	"webtyp.com/css"
	"webtyp.com/widget"
	"webtyp.com/widget/style"
)

// RenderCSS lays the chips in a row that wraps, packed to the start: each chip
// is as wide as its label ("09:30" or "Traumatología"), so a long label is
// never clipped and three chips do not spread across the whole width. The
// picked chip takes the primary surface, a disabled one the inactive look.
func (g *ChipGrid) RenderCSS() *css.Stylesheet {
	return style.For(g).
		Root(style.Row(style.Space2)).
		Part(PartChip,
			style.Button(style.Panel),
			style.Round(style.RadiusMd),
		).
		Part(PartEmpty,
			style.FontSize(style.TextSm),
		).
		When(widget.Selected, PartChip, style.As(style.Primary)).
		When(widget.Disabled, PartChip, style.As(style.Inactive)).
		Stylesheet()
}
