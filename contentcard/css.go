//go:build !wasm

package contentcard

import (
	"webtyp.com/css"
	"webtyp.com/widget/style"
)

// RenderCSS defines the visual sheet for contentcard.
func (c *ContentCard) RenderCSS() *css.Stylesheet {
	return style.For(c).
		Root(
			style.Stack(style.SpaceNone),
			style.As(style.Panel),
			style.Round(style.RadiusMd),
			style.Raise(style.Raised),
		).
		Part(PartHeader,
			style.Pad(style.Space2),
			style.Heading(),
		).
		Part(PartBody,
			style.Pad(style.Space2),
			style.Fill(),
		).
		Part(PartFooter,
			style.Pad(style.Space2),
			style.As(style.Inset),
		).
		Stylesheet()
}
