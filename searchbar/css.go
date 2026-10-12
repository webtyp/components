//go:build !wasm

package searchbar

import (
	"webtyp.com/css"
	"webtyp.com/widget"
	"webtyp.com/widget/style"
)

// RenderCSS defines the search bar's visual contract using the style DSL.
// The searchbar integrates subtle magnifier and bare input into one cohesive field.
func (s *SearchBar) RenderCSS() *css.Stylesheet {
	return style.For(s).
		Root(
			style.Row(style.SpaceNone),
			style.Round(style.RadiusMd),
			style.HideOverflow(),
			style.ControlBox(),
			style.KeepSize(),
			style.As(style.Field),
		).
		On(css.Mobile, widget.Part(""),
			style.ControlBoxCompact(),
		).
		Part(PartIcon,
			style.As(style.Subtle),
			style.IconCap(),
		).
		On(css.Mobile, PartIcon,
			style.IconCapCompact(),
		).
		Part(PartInput,
			style.As(style.Bare),
			style.PadInline(style.Space2),
			style.Grow(),
		).
		Stylesheet()
}
