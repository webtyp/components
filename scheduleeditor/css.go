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
		When(widget.Invalid, PartPatternRow,
			style.As(style.DangerWash),
		).
		Part(PartDayChips,
			style.Row(style.Space1),
		).
		Part(PartDayChip,
			style.ControlBox(),
			style.Round(style.RadiusSm),
			style.Interactive(style.Subtle),
		).
		Part(PartRowRemove,
			style.ControlBox(),
			style.Interactive(style.Danger),
			style.Round(style.RadiusSm),
			style.KeepSize(),
		).
		Part(PartRowAdd,
			style.ControlBox(),
			style.Interactive(style.Primary),
			style.Round(style.RadiusSm),
			style.KeepSize(),
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
			style.ControlBox(),
			style.Interactive(style.Primary),
			style.Round(style.RadiusSm),
			style.KeepSize(),
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
			style.ControlBox(),
			style.Interactive(style.Danger),
			style.Round(style.RadiusSm),
			style.KeepSize(),
		)
}
