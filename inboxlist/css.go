//go:build !wasm

package inboxlist

import (
	"webtyp.com/css"
	"webtyp.com/widget"
	"webtyp.com/widget/style"
)

// RenderCSS defines the inboxlist visual contract using the style DSL.
func (l *InboxList) RenderCSS() *css.Stylesheet {
	return style.For(l).
		Root(
			style.Fill(),
			style.Stack(style.SpaceNone),
		).
		Part(PartList,
			style.Stack(style.Space1),
		).
		Part(PartRow,
			style.Anchor(),
			style.Button(style.Page),
			style.Pad(style.Space3),
			style.Round(style.RadiusMd),
			style.Stack(style.Space1),
		).
		Part(PartTitle,
			style.FontSize(style.TextSm),
			style.FontWeight(style.WeightMedium),
		).
		Part(PartTitleUnread,
			style.FontSize(style.TextSm),
			style.FontWeight(style.WeightBold),
		).
		Part(PartTime,
			style.FontSize(style.TextXs),
			style.As(style.Subtle),
		).
		Part(PartPreview,
			style.FontSize(style.TextXs),
			style.As(style.Subtle),
		).
		Part(PartEmpty,
			style.Pad(style.Space4),
			style.As(style.Subtle),
			style.CenterContent(),
		).
		When(widget.Selected, PartRow,
			style.As(style.AccentWash),
		).
		Stylesheet()
}
