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
			style.CrossCenter(),
			// Sunken background
			css.Set("background-color", "var(--color-surface-sunken)"), // Unconfirmed but keeping per plan
			css.Set("border-radius", "var(--radius-md)"),
			css.Set("padding", "var(--space-1)"),
		).
		Part(PartPill,
			style.Button(style.Surface),
			css.Set("background-color", "transparent"),
			css.Set("border-radius", "var(--radius-sm)"),
			css.Set("color", "var(--color-muted)"),
			// Raised active pill
			style.When(widget.Selected,
				css.Set("background-color", "var(--color-surface)"),
				css.Set("color", "var(--color-on-surface)"),
				css.Set("box-shadow", "var(--shadow-sm)"),
			),
		).
		Stylesheet()
}
