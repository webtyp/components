//go:build !wasm

package agedate

import (
	"webtyp.com/css"
	"webtyp.com/widget/style"
)

func (a *AgeDate) RenderCSS() *css.Stylesheet {
	s := style.For(a)

	s.Root(
		style.Row(style.Space2),
	)

	s.Part(PartAge,
		style.ControlBox(),
		style.As(style.Field),
		style.Round(style.RadiusMd),
		style.PadInline(style.Space4),
	)

	s.Part(PartDate,
		style.ControlBox(),
		style.As(style.Field),
		style.Round(style.RadiusMd),
		style.PadInline(style.Space4),
		style.Grow(),
	)

	s.Part(PartError,
		style.Glyph(style.Danger),
		style.FontSize(style.TextXs),
	)

	return s.Stylesheet()
}
