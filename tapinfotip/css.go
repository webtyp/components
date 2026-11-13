//go:build !wasm

package tapinfotip

import (
	"webtyp.com/css"
	"webtyp.com/widget"
	"webtyp.com/widget/style"
)

// RenderCSS defines the tapinfotip visual contract. The button is a 44px touch
// target around a quiet glyph (a help affordance, not a call to action); the
// panel is small, muted text that appears only while open.
func (t *TapInfotip) RenderCSS() *css.Stylesheet {
	return style.For(t).
		Root(
			style.Stack(style.Space1),
		).
		// The (i) is as tall as the label text beside it (1em of the button's
		// own font size, the label's size), sits at the end of the line and is
		// coloured quietly (Subtle: the muted text colour, which adapts to light and dark).
		Part(PartButton,
			style.Interactive(style.Bare),
			style.As(style.Subtle),
			style.Round(style.RadiusFull),
			style.FontSize(style.TextSm),
			style.PushEnd(),
		).
		// A bare glyph is not a box: hover and press recolour the glyph itself
		// (muted → accent) and draw no ring or fill; keyboard focus keeps the
		// accent ring so it stays visible (ui-consistency, «Icon buttons»).
		Cue(widget.Hover, PartButton, style.Glyph(style.Accent)).
		Cue(widget.Press, PartButton, style.Glyph(style.Accent)).
		Cue(widget.Focus, PartButton, style.As(style.AccentOutline)).
		Part(PartIcon,
			style.IconBox(style.IconSm),
		).
		Part(PartPanel,
			style.Glyph(style.Inactive),
			style.FontSize(style.TextXs),
			style.RevealedBy(widget.Open),
		).
		Stylesheet()
}
