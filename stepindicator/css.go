//go:build !wasm

package stepindicator

import (
	"webtyp.com/css"
	"webtyp.com/widget"
	"webtyp.com/widget/style"
)

// RenderCSS returns the stylesheet for the stepindicator.
func (c *StepIndicator) RenderCSS() *css.Stylesheet {
	return style.For(c).
		Root(
			style.Row(style.Space4),
			style.CenterContent(),
			style.KeepSize(),
		).
		Part(PartStep,
			style.Button(style.Bare),
			style.Row(style.Space2),
			style.CenterContent(),
			style.KeepSize(),
		).
		Part(PartBadge,
			style.IconBox(style.IconMd),
			style.CenterContent(),
			style.Round(style.RadiusFull),
			style.As(style.Panel),
			style.FontSize(style.TextSm),
		).
		Part(PartLabel,
			style.FontSize(style.TextSm),
		).
		Part(PartLine,
			style.Grow(),
			style.DividerBelow(),
		).
		When(widget.Current, PartStep,
			style.FontWeight(style.WeightBold),
		).
		WhenWithin(widget.Current, PartStep, PartBadge,
			style.As(style.Primary),
		).
		WhenWithin(widget.Selected, PartStep, PartBadge,
			style.As(style.AccentInverse),
		).
		Stylesheet()
}
