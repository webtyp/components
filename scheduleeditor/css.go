//go:build !wasm

package scheduleeditor

import (
	"webtyp.com/css"
	"webtyp.com/widget"
	"webtyp.com/widget/style"
)

// RenderCSS implements the visual contract for ScheduleEditor.
func (e *ScheduleEditor) RenderCSS() *css.Stylesheet {
	return e.sheet().Stylesheet()
}

func (e *ScheduleEditor) sheet() *style.Sheet {
	return style.For(e).
		Root(
			style.Stack(style.Space3),
			style.Fill(),
		).
		Part(PartPattern,
			style.Stack(style.Space2),
		).
		Part(PartPatternRow,
			style.Row(style.Space2),
			style.ControlBox(),
			style.Round(style.RadiusMd),
			style.Anchor(),
		).
		When(widget.Selected, PartDayLabel,
			style.As(style.Primary),
		).
		When(widget.Invalid, PartPatternRow,
			style.As(style.DangerWash),
		).
		Part(PartDayChips,
			style.Row(style.Space1),
		).
		// The checkbox itself: the real control, out of sight but still
		// focusable and announced. A native checkbox brings its own skin and
		// cannot be painted, so the visible pill is its <label> below.
		Part(PartDayChip,
			style.VisuallyHidden(),
		).
		// The pill. Button() because that is what it is — something the user
		// presses — which also gives it the 44px control box the tap-target
		// floor needs, now that the input no longer carries it.
		Part(PartDayLabel,
			style.Button(style.Subtle),
		).
		Part(PartFieldLabel,
			style.As(style.Subtle),
			style.FontSize(style.TextSm),
		).
		// Subtle, not Danger: one of these repeats on every pattern row, and a
		// column of red blocks reads as an alarm rather than a control. The
		// danger tint arrives on hover, from Interactive's own derivation.
		Part(PartRowRemove,
			style.Button(style.Subtle),
		).
		Part(PartRowAdd,
			style.Button(style.Primary),
		).
		Part(PartMarker,
			style.Stack(style.Space2),
			style.As(style.Panel),
			style.Round(style.RadiusMd),
			style.Pad(style.Space3),
		).
		Part(PartMarkerHours,
			style.Row(style.Space2),
		).
		Part(PartExceptions,
			style.Stack(style.Space2),
		).
		Part(PartSlider,
			style.As(style.Panel),
			style.Round(style.RadiusMd),
			style.Pad(style.Space3),
		).
		Part(PartExcForm,
			style.RevealedBy(widget.Open),
			style.Stack(style.Space2),
			style.As(style.Inset),
			style.Round(style.RadiusMd),
			style.Pad(style.Space3),
		).
		Part(PartExcHours,
			style.RevealedBy(widget.Open),
			style.Row(style.Space2),
		).
		Part(PartExcType,
			style.Row(style.Space2),
			style.ControlBox(),
		).
		Part(PartExcNotes,
			style.ControlBox(),
			style.Round(style.RadiusSm),
			style.As(style.Inset),
		).
		Part(PartExcAdd,
			style.Button(style.Primary),
		).
		Part(PartExcList,
			style.Stack(style.Space1),
		).
		Part(PartExcItem,
			style.Row(style.Space2),
			style.ControlBox(),
			style.Round(style.RadiusMd),
			style.As(style.Panel),
		).
		Part(PartExcHoliday,
			style.As(style.Subtle),
		).
		Part(PartExcRemove,
			style.Button(style.Danger),
		)
}
