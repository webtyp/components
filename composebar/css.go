//go:build !wasm

package composebar

import (
	"webtyp.com/css"
	"webtyp.com/widget/style"
)

// RenderCSS defines the composebar visual contract using the style DSL.
func (c *ComposeBar) RenderCSS() *css.Stylesheet {
	return style.For(c).
		Root(
			style.Row(style.Space2),
			style.Pad(style.Space2),
			style.As(style.Panel),
		).
		Part(PartInput,
			style.Grow(),
			style.ControlBox(),
			style.Pad(style.Space2),
			style.Round(style.RadiusMd),
		).
		Part(PartSend,
			style.Button(style.Primary),
		).
		Stylesheet()
}
