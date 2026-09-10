//go:build !wasm

package targetdate

import (
	"webtyp.com/components/listgap"
	"webtyp.com/components/listselect"
	"webtyp.com/css"
	"webtyp.com/widget"
	"webtyp.com/widget/style"
)

// RenderCSS defines the targetdate visual contract using the style DSL.
func (t *TargetDate) RenderCSS() *css.Stylesheet {
	return t.sheet().Stylesheet()
}

// sheet builds the style Sheet.
func (t *TargetDate) sheet() *style.Sheet {
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
		Part(PartContent,
			style.Row(style.Space2),
			style.Grow(),
			style.Pad(style.Space2),
		).
		Part(PartLead,
			style.StartContent(),
			style.MediaBox(style.AspectSquare),
			style.ControlBox(),
			style.Pad(style.Space2),
			style.Divider(style.SideEnd),
			style.KeepSize(),
		).
		Part(PartLeadStack,
			style.Stack(style.SpaceNone),
			style.StartContent(),
		).
		Part(PartLeadTop,
			style.FontSize(style.TextXs),
		).
		Part(PartLeadMain,
			style.FontSize(style.TextSm),
			style.FontWeight(style.WeightBold),
		).
		Part(PartLeadBottom,
			style.FontSize(style.TextXs),
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
