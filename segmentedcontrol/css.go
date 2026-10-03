//go:build !wasm

package segmentedcontrol

import (
	"webtyp.com/css"
	"webtyp.com/widget"
	"webtyp.com/widget/style"
)

func (c *SegmentedControl) RenderCSS() *css.Stylesheet {
	return style.For(c).
		Root(
			style.Row(style.Space1),
			style.CenterContent(),
			style.As(style.Inset),
			style.Round(style.RadiusMd),
			style.Pad(style.Space1),
		).
		Part(PartPill,
			style.Button(style.Bare),
			style.Round(style.RadiusSm),
			style.FontSize(style.TextSm),
		).
		When(widget.Selected, PartPill,
			style.As(style.Panel),
			style.Raise(style.Raised),
		).
		Stylesheet()
}
