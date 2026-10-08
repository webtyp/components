//go:build !wasm

package statgrid

import (
	"webtyp.com/css"
	"webtyp.com/widget/style"
)

// RenderCSS defines the stylesheet for statgrid.
func (s *StatGrid) RenderCSS() *css.Stylesheet {
	return style.For(s).
		Root(
			style.Grid(4, style.ColumnMedium, style.Space4),
		).
		Part(PartItem,
			style.Stack(style.Space1),
			style.As(style.Panel),
			style.Pad(style.Space4),
			style.Round(style.RadiusMd),
			style.CenterContent(),
		).
		Part(PartValue,
			style.FontSize(style.Text2xl),
			style.FontWeight(style.WeightBold),
		).
		Part(PartLabel,
			style.FontSize(style.TextSm),
		).
		Stylesheet()
}
