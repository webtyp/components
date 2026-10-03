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
			style.Row(style.Space4), // Using valid layout token Space4 based on "have (), want (style.Space)"
			style.CrossCenter(),
			style.KeepSize(),
		).
		Part(PartStep,
			style.Button(style.Surface),
			style.Row(style.Space2),
			style.CrossCenter(),
		).
		Part(PartBadge,
			style.CenterContent(),
			css.Set("width", "1.5rem"),
			css.Set("height", "1.5rem"),
			css.Set("border-radius", "var(--radius-full)"),
			css.Set("border", "1px solid var(--color-outline)"),
			css.Set("background-color", "var(--color-surface)"),
			css.Set("color", "var(--color-muted)"),
			css.Set("font-size", "var(--text-sm)"),
			css.Set("font-weight", "var(--font-weight-medium)"),
		).
		Part(PartLabel,
			css.Set("color", "var(--color-on-surface)"),
			css.Set("white-space", "nowrap"),
		).
		Part(PartLine,
			style.Grow(),
			css.Set("height", "1px"),
			css.Set("background-color", "var(--color-outline)"),
		).
		// styling for active step
		Rule(
			css.Class(string(clsStep)).When(widget.Current).Descendant(css.Class(string(clsLabel))),
			css.Set("font-weight", "var(--font-weight-bold)"),
		).
		Rule(
			css.Class(string(clsStep)).When(widget.Current).Descendant(css.Class(string(clsBadge))),
			css.Set("background-color", "var(--color-primary)"),
			css.Set("color", "var(--color-on-primary)"),
			css.Set("border-color", "var(--color-primary)"),
		).
		// styling for done step
		Rule(
			css.Class(string(clsStep)).When(widget.Selected).Descendant(css.Class(string(clsBadge))),
			css.Set("background-color", "var(--color-primary)"),
			css.Set("color", "var(--color-on-primary)"),
			css.Set("border-color", "var(--color-primary)"),
		).
		Rule(
			css.Class(string(clsLine)).When(widget.Current),
			css.Set("background-color", "var(--color-primary)"),
		).
		Stylesheet()
}
