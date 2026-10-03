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
			style.Fill(),
		).
		Part(PartCell,
			style.Button(style.Bare),
			style.CenterContent(),
		).
		When(widget.Selected, PartCell,
			style.As(style.Primary),
		).
		Stylesheet()
}
