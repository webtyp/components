//go:build !wasm

package chipgrid

import (
	"webtyp.com/css"
	"webtyp.com/widget"
	"webtyp.com/widget/style"
)

// RenderCSS lays the chips in a grid that reflows by container width; the
// picked chip takes the primary surface, a disabled one the inactive look.
func (g *ChipGrid) RenderCSS() *css.Stylesheet {
	return style.For(g).
		Root(style.Grid(6, style.ColumnNarrow, style.Space2)).
		Part(PartChip,
			style.Button(style.Panel),
			style.ChipBox(),
			style.Round(style.RadiusMd),
		).
		Part(PartEmpty,
			style.SpanFull(),
			style.FontSize(style.TextSm),
		).
		When(widget.Selected, PartChip, style.As(style.Primary)).
		When(widget.Disabled, PartChip, style.As(style.Inactive)).
		Stylesheet()
}
