//go:build !wasm

package targethour

import (
	"webtyp.com/components/listgap"
	"webtyp.com/components/listselect"
	"webtyp.com/css"
	"webtyp.com/widget"
	"webtyp.com/widget/style"
)

// RenderCSS defines the targethour visual contract using the style DSL.
func (t *TargetHour) RenderCSS() *css.Stylesheet {
	return t.sheet().Stylesheet()
}

// sheet builds the style Sheet.
func (t *TargetHour) sheet() *style.Sheet {
	s := style.For(t).Root(
		style.Fill(),
		style.Stack(style.SpaceNone),
	)
	listgap.Apply(s, PartList)
	s.On(css.Mobile, PartList, listgap.MobileOpts()...)
	listselect.ApplyRow(s, PartRow)
	listselect.ApplyHeader(s)

	return s.
		Part(PartRow,
			style.Anchor(),
			style.Row(style.Space2),
			style.KeepSize(),
			style.ControlBox(),
			style.As(style.Page),
			style.Interactive(style.Page),
			style.Round(style.RadiusMd),
		).
		Part(PartFree,
			style.Anchor(),
			style.Row(style.Space2),
			style.KeepSize(),
			style.ControlBox(),
			style.Interactive(style.Page),
			style.Round(style.RadiusMd),
			style.As(style.Subtle),
			// Dashed leading edge: visually a "still to fill" row, never a
			// fully booked record. Kept as a side inset so the row keeps its
			// Interactive(Page) surface over the whole box.
			style.PadInline(style.Space3),
		).
		Part(PartFreeAdd,
			style.FontSize(style.TextLg),
			style.FontWeight(style.WeightBold),
			style.CenterContent(),
			style.KeepSize(),
		).
		Part(PartContent,
			style.Row(style.Space2),
			style.Grow(),
			style.Pad(style.Space2),
		).
		Part(PartHour,
			style.FontSize(style.TextLg),
			style.FontWeight(style.WeightBold),
			style.KeepSize(),
			style.CenterContent(),
			style.Divider(style.SideEnd),
			style.Pad(style.Space2),
		).
		Part(PartLabel,
			style.FontWeight(style.WeightBold),
			style.Grow(),
		).
		Part(PartBadge,
			style.As(style.Inset),
			style.Round(style.RadiusSm),
			style.FontSize(style.TextXs),
			style.ChipBox(),
			style.CenterContent(),
			style.KeepSize(),
			style.OnEdge(style.EdgeBottom, style.SideEnd, style.SpaceNone, style.Space3),
		).
		When(widget.Selected, PartRow,
			style.As(style.Accent),
		).
		When(widget.Locked, PartRow,
			style.As(style.AccentWash),
		).
		When(widget.Busy, PartRow,
			style.As(style.Subtle),
		).
		Cue(widget.Hover, PartRow,
			style.As(style.AccentWash),
		).
		Cue(widget.Focus, PartRow,
			style.As(style.AccentWash),
		).
		Cue(widget.Press, PartRow,
			style.As(style.Accent),
		)
}
