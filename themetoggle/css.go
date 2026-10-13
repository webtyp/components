//go:build !wasm

package themetoggle

import (
	"webtyp.com/css"
	"webtyp.com/widget"
	"webtyp.com/widget/style"
)

// RenderCSS defines the themetoggle visual contract using style DSL.
func (t *ThemeToggle) RenderCSS() *css.Stylesheet {
	return style.For(t).
		Root(
			style.Row(style.Space1),
			style.As(style.Inset),
			style.Round(style.RadiusFull),
			style.Pad(style.Space1),
		).
		Part(PartOption,
			style.Button(style.Subtle),
			style.Round(style.RadiusFull),
			style.Pad(style.Space1),
			style.IconBox(style.IconSm),
		).
		Part(PartOption, style.RevealedBy(widget.Current), style.As(style.Panel)).
		Stylesheet()
}
