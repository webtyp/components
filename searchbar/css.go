//go:build !wasm

package searchbar

import (
	"webtyp.com/css"
	"webtyp.com/widget/style"
)

// RenderCSS defines the search bar's visual contract using the style DSL.
func (s *SearchBar) RenderCSS() *css.Stylesheet {
	return style.For(s).
		// The bar is ONE control, not a card holding two loose pieces: the
		// magnifier is the bar's leading cap, the input its body, and a gap or a
		// card of its own between them saws the bar back into separate boxes.
		// The root carries the radius and clips, so cap and body stay square and
		// still read as one rounded bar.
		// ControlBox pins the bar to --control-height, the token every control
		// in the ecosystem answers to — that is what lets a host stack this bar
		// against its own buttons and have the heights agree by construction.
		Root(
			style.Row(style.SpaceNone),
			style.Round(style.RadiusMd),
			style.HideOverflow(),
			style.ControlBox(),
			style.KeepSize(),
		).
		// The magnifier is the bar's square cap: IconCap, not padding, sets the
		// width — a padded box drifts off the control token (the old
		// Pad(Space2)+icon-box measured 40px against the host's 66), while the
		// square derives from the same --control-height as everything else.
		// IconCap also sizes PartGlyph's <svg> (half the cap), which is why
		// that part carries no IconBox of its own: the four options this used
		// to spell out by hand are the recipe now, and so is the glyph size
		// the four never covered.
		Part(PartIcon,
			style.As(style.Primary),
			style.IconCap(),
		).
		// The input is the body of the bar: it grows into whatever the cap
		// leaves and answers to the same control height, so cap and body can
		// never drift apart vertically — the mismatch that left a 25px field
		// floating in the middle of a 72px strip.
		Part(PartInput,
			style.As(style.Inset),
			style.Pad(style.Space2),
			style.Grow(),
			style.ControlBox(),
		).
		Stylesheet()
}
