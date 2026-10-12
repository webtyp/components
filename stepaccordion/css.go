//go:build !wasm

package stepaccordion

import (
	"webtyp.com/css"
	"webtyp.com/widget"
	"webtyp.com/widget/style"
)

// RenderCSS lays the steps out as a column of panels. What each step shows
// hangs off its phase: the body only while open, summary and Edit only once
// finished (widget.Locked). A locked step (widget.Disabled) shows its number
// and title, dimmed.
func (a *StepAccordion) RenderCSS() *css.Stylesheet {
	return style.For(a).
		Root(style.Stack(style.Space3)).
		Part(PartStep,
			style.Stack(style.Space3),
			style.As(style.Panel),
			style.Round(style.RadiusMd),
			style.Pad(style.Space4),
		).
		Part(PartHeader,
			style.Row(style.Space3),
			style.CenterContent(),
		).
		Part(PartBadge,
			style.ChipBox(),
			style.As(style.AccentWash),
			style.Round(style.RadiusFull),
			style.FontWeight(style.WeightBold),
		).
		Part(PartTitle,
			style.FontSize(style.TextBase),
			style.FontWeight(style.WeightMedium),
		).
		Part(PartSummary,
			style.FontSize(style.TextSm),
			style.PushEnd(),
			style.Hide(),
		).
		Part(PartEdit,
			style.Button(style.Bare),
			style.Hide(),
		).
		Part(PartBody, style.Hide()).
		WhenWithin(widget.Open, PartStep, PartBody, style.Show()).
		WhenWithin(widget.Locked, PartStep, PartSummary, style.Show()).
		WhenWithin(widget.Locked, PartStep, PartEdit, style.Show()).
		When(widget.Disabled, PartStep, style.As(style.Inactive)).
		Stylesheet()
}
