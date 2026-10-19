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
		// The badge is data (a code), not an action: neutral and outlined, never
		// the primary fill, which belongs to action buttons. It rides the row's
		// bottom-right edge, out of the flow (OnEdge), so the label always has
		// the row's whole width and the badge never moves — it does not wrap
		// under the label when the two do not fit side by side.
		Part(PartBadge,
			style.As(style.Panel),
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
		// Outline for a passing state, fill for a confirmed one: hover and focus
		// outline the row in the accent and leave its content alone; pressed and
		// selected fill it.
		Cue(widget.Hover, PartRow,
			style.As(style.AccentOutline),
		).
		Cue(widget.Focus, PartRow,
			style.As(style.AccentOutline),
		).
		Cue(widget.Press, PartRow,
			style.As(style.Accent),
		)
}
