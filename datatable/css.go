//go:build !wasm

package datatable

import (
	"webtyp.com/css"
	"webtyp.com/widget/style"
)

// RenderCSS defines the datatable visual contract using the style DSL.
func (t *DataTable) RenderCSS() *css.Stylesheet {
	return style.For(t).
		Root(
			style.Width(style.Full),
			style.As(style.Panel),
			style.Round(style.RadiusNone),
		).
		Part(PartHeader,
			style.As(style.Inset),
			style.FontWeight(style.WeightBold),
			style.Pad(style.Space2),
			style.Round(style.RadiusNone),
		).
		Part(PartRow,
			style.Pad(style.Space2),
			style.As(style.Panel),
			style.Interactive(style.Panel),
			style.Round(style.RadiusNone),
		).
		Stylesheet()
}
