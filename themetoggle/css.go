//go:build !wasm

package themetoggle

import (
	"webtyp.com/css"
	"webtyp.com/widget/style"
)

// RenderCSS defines the themetoggle visual contract using style DSL. It is a
// "Fixed floating button" (see README) — Docked(Viewport, ...) is what makes
// that true; a consumer must not have to reposition it itself with inline
// CSS, that would be exactly the hardcoded-style workaround the harness
// forbids.
func (t *ThemeToggle) RenderCSS() *css.Stylesheet {
	return style.For(t).
		Root(
			style.Button(style.Primary),
			// Round survives the Button migration: this control is a circle, and
			// Button only derives its surface's default radius (RadiusSm).
			style.Round(style.RadiusFull),
			style.Pad(style.Space1),
			style.Docked(style.Viewport, style.EdgeTop, style.SideEnd, style.Space4),
		).
		Stylesheet()
}
