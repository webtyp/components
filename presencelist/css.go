//go:build !wasm

package presencelist

import (
	"webtyp.com/css"
	"webtyp.com/widget/style"
)

// RenderCSS defines the presencelist visual contract using the style DSL.
func (l *PresenceList) RenderCSS() *css.Stylesheet {
	return style.For(l).
		Root(
			style.Fill(),
			style.Stack(style.SpaceNone),
		).
		Part(PartList,
			style.Stack(style.Space1),
		).
		Part(PartRow,
			style.Button(style.Page),
			style.Row(style.Space2),
			style.Pad(style.Space2),
			style.Round(style.RadiusMd),
			style.CenterContent(),
		).
		Part(PartDotOnline,
			style.IconBox(style.IconSm),
			style.Round(style.RadiusFull),
			style.As(style.Accent),
		).
		Part(PartDotOffline,
			style.IconBox(style.IconSm),
			style.Round(style.RadiusFull),
			style.As(style.Subtle),
		).
		Part(PartLabel,
			style.Grow(),
			style.FontSize(style.TextSm),
		).
		Part(PartStatus,
			style.VisuallyHidden(),
		).
		Part(PartEmpty,
			style.Pad(style.Space4),
			style.As(style.Subtle),
			style.CenterContent(),
		).
		Stylesheet()
}
