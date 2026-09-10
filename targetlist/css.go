//go:build !wasm

package targetlist

import (
	"webtyp.com/components/listgap"
	"webtyp.com/components/listselect"
	"webtyp.com/css"
	"webtyp.com/widget"
	"webtyp.com/widget/style"
)

// RenderCSS defines the targetlist visual contract using the style DSL.
func (t *TargetList) RenderCSS() *css.Stylesheet {
	return t.sheet().Stylesheet()
}

// sheet builds the style Sheet.
func (t *TargetList) sheet() *style.Sheet {
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
			style.Pad(style.Space3),
			style.Round(style.RadiusMd),
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
