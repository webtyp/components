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
		// Split, no Stack: el editor tiene exactamente dos mitades — el patrón
		// semanal y las fechas específicas — y Split las pone lado a lado
		// cuando hay ancho y las apila cuando no, que es la definición de la
		// receta ("two panels that stack below their own width"). Apiladas
		// siempre, el patrón ocupaba una columna angosta y el resto del
		// desktop quedaba vacío, mientras el calendario caía tan abajo que
		// había que scrollear para saber que existía.
		//
		// TwoThirds al patrón: es el que tiene siete filas con selectores; el
		// calendario es un bloque de ancho fijo y no gana nada con más sitio.
		Root(
			style.Split(style.SplitTwoThirds, style.Space3),
			style.Fill(),
		).
		Part(PartPattern,
			style.Stack(style.Space2),
		).
		// La fila de UN día: nombre, interruptor y sus rangos. PadInline
		// porque lleva radio y un relleno de estado (Invalid), y contenido
		// pegado a un borde redondeado se lee como un recorte, no como tarjeta.
		Part(PartDayRow,
			style.Row(style.Space2),
			style.ControlBox(),
			style.Round(style.RadiusMd),
			style.PadInline(style.Space2),
			style.Anchor(),
		).
		// Los rangos del día, a la derecha del nombre. Grow para que el nombre
		// quede a la izquierda y los horarios ocupen el resto.
		Part(PartDaySlots,
			style.Row(style.Space2),
			style.Grow(),
		).
		// "No trabaja" es texto apagado: informa sin competir con los días que
		// sí tienen horario.
		Part(PartDayOffText,
			style.As(style.Subtle),
			style.FontSize(style.TextSm),
			style.Grow(),
		).
		// La ayuda de una línea bajo cada título. Existe porque la pantalla
		// tiene que decir qué hacer sin que nadie la explique.
		Part(PartHintText,
			style.As(style.Subtle),
			style.FontSize(style.TextSm),
		).
		// "+" y "×" son controles de un carácter: caja de icono, no de botón
		// con etiqueta.
		Part(PartSlotAdd,
			style.Button(style.Inset),
		).
		Part(PartSlotRemove,
			style.Button(style.Inset),
		).
		// The time range travels as one block: KeepSize forbids the wrap
		// INSIDE it, which is the whole reason it exists.
		Part(PartTimeRange,
			style.Row(style.Space1),
			style.CenterContent(),
			style.KeepSize(),
		).
		// A section heading, not body text. The three sections were bare
		// <div>s carrying plain paragraph type, so "Weekly pattern" and
		// "Marked days" read as stray sentences rather than as the titles of
		// the panels underneath them.
		Part(PartSectionTitle,
			style.FontSize(style.TextLg),
			style.FontWeight(style.WeightBold),
		).
		When(widget.Selected, PartDayLabel,
			style.As(style.Primary),
		).
		// El estado inválido vive en la fila del DÍA, que es donde el
		// solapamiento ocurre ahora: dos rangos del mismo día que se pisan.
		When(widget.Invalid, PartDayRow,
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
		// Inset, not Subtle: Subtle paints nothing, so an unchosen day read as
		// loose text beside the chosen ones instead of as a pill you can
		// press. Inset gives it the sunken fill and outline that says control.
		// El nombre del día es la píldora que enciende y apaga el día. KeepSize
		// y un ancho propio para que los siete nombres formen una columna: sin
		// eso, "Miércoles" y "Jue" empujan sus horarios a distinta altura y la
		// lista deja de leerse como una tabla.
		Part(PartDayLabel,
			style.Button(style.Inset),
			style.ChipBox(),
			style.KeepSize(),
		).
		Part(PartFieldLabel,
			style.As(style.Subtle),
			style.FontSize(style.TextSm),
		).
		// Inset, not Danger: one of these repeats on every pattern row, and a
		// column of red blocks reads as an alarm rather than a control. Inset
		// keeps it legible as a button; the tint arrives on hover.
		Part(PartRowRemove,
			style.Button(style.Inset),
			style.PushEnd(),
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
		// Cada acción es un bloque que se revela solo cuando el día elegido la
		// admite. Tres bloques, de los que como mucho uno está abierto: el
		// estado del día es lo que decide, no el usuario.
		Part(PartExcAction,
			style.RevealedBy(widget.Open),
			style.Stack(style.Space2),
		).
		// La línea que explica por qué se ofrece esta acción y no otra.
		Part(PartExcHint,
			style.As(style.Subtle),
			style.FontSize(style.TextSm),
		).
		Part(PartExcDateRow,
			style.Row(style.Space2),
			style.CenterContent(),
		).
		// Sin RevealedBy: el bloque de acción que las contiene es quien revela.
		// El reveal que había acá venía del formulario viejo, donde la fila de
		// horas se ocultaba al elegir "Cerrado"; con las acciones contextuales
		// quedó como una segunda compuerta que nada abría, y los selectores de
		// horario eran invisibles para siempre.
		Part(PartExcHours,
			style.Row(style.Space2),
			style.CenterContent(),
		).
		// La fecha abre la fila y es su ancla de lectura: negrita y sin
		// encogerse, para que la columna de fechas quede alineada aunque los
		// tipos y las notas tengan largos distintos.
		Part(PartExcDate,
			style.FontWeight(style.WeightBold),
			style.KeepSize(),
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
