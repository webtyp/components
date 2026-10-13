//go:build !wasm

package themetoggle

import (
	"webtyp.com/css"
	"webtyp.com/widget"
	"webtyp.com/widget/style"
)

// RenderCSS defines the themetoggle visual contract using style DSL. It has no
// position of its own: where it sits is the host's decision (platformd puts it
// in the user menu).
func (t *ThemeToggle) RenderCSS() *css.Stylesheet {
	return style.For(t).
		Root(
			style.Row(style.Space1),
			style.As(style.Inset),
			style.Round(style.RadiusFull),
			style.Pad(style.Space1),
		).
		// Quiet segments: the control sits inside a menu panel and must not
		// outshine the content around it.
		Part(PartOption,
			style.Button(style.Subtle),
			style.Round(style.RadiusFull),
			style.Pad(style.Space1),
		).
		Part(PartIcon, style.IconBox(style.IconSm)).
		// The choice in effect is a raised segment, never a hidden sibling.
		When(widget.Current, PartOption, style.As(style.Panel)).
		Stylesheet()
}
