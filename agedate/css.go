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
		style.As(style.Page),
	)

	s.Part(PartDate,
		style.ControlBox(),
		style.As(style.Page),
		style.Grow(),
	)

	return s.Stylesheet()
}
