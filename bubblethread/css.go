//go:build !wasm

package bubblethread

import (
	"webtyp.com/css"
	"webtyp.com/widget/style"
)

// RenderCSS defines the bubblethread visual contract using the style DSL.
func (t *BubbleThread) RenderCSS() *css.Stylesheet {
	return style.For(t).
		Root(
			style.Fill(),
			style.Stack(style.SpaceNone),
		).
		Part(PartThread,
			style.Scroll(),
			style.Fill(),
			style.Pad(style.Space3),
			style.Stack(style.Space3),
		).
		Part(PartMine,
			style.PushEnd(),
			style.As(style.AccentWash),
			style.Pad(style.Space3),
			style.Round(style.RadiusLg),
			style.Stack(style.Space1),
		).
		Part(PartTheirs,
			style.As(style.Inset),
			style.Pad(style.Space3),
			style.Round(style.RadiusLg),
			style.Stack(style.Space1),
		).
		Part(PartAuthor,
			style.FontSize(style.TextXs),
			style.FontWeight(style.WeightBold),
			style.As(style.Subtle),
		).
		Part(PartBody,
			style.FontSize(style.TextSm),
		).
		Part(PartTime,
			style.FontSize(style.TextXs),
			style.As(style.Subtle),
		).
		Part(PartRead,
			style.FontSize(style.TextXs),
			style.As(style.Subtle),
		).
		Part(PartEmpty,
			style.Pad(style.Space4),
			style.As(style.Subtle),
			style.CenterContent(),
		).
		Stylesheet()
}
