//go:build !wasm

package cellgrid

import (
	"webtyp.com/css"
	"webtyp.com/widget"
	"webtyp.com/widget/style"
)

func (c *CellGrid) RenderCSS() *css.Stylesheet {
	return style.For(c).
		Root(
			css.Set("display", "grid"),
		).
		Part(PartCell,
			css.Set("aspect-ratio", "1 / 1"),
			css.Set("border", "1px solid var(--color-outline)"),
			style.When(widget.Selected,
				css.Set("background-color", "var(--color-primary)"),
			),
		).
		Stylesheet()
}
