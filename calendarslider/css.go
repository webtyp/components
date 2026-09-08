//go:build !wasm

package calendarslider

import (
	"webtyp.com/css"
	"webtyp.com/widget"
	"webtyp.com/widget/style"
)

// RenderCSS describe el aspecto del calendario con los tokens del tema. La
// puesta en escena es fiel al calendario original: un mes a la vez, centrado
// en la página (Center), con las secciones vecinas fuera de vista en una
// tira de scroll-snap (ScrollRow) — cada mes ocupa el 100% de la tira, así
// que solo uno es visible; el hoy se rellena con el color primario y las
// flechas ‹ › son una fila compacta centrada bajo el nombre del mes
// (PartMonthNav), idéntica en desktop y táctil. Es estática a propósito:
// sin reveal en hover/foco — volver las flechas absolutas las hacía
// desaparecer bajo el puntero, y ocultar la fila con foco cancelaba los
// toques en mobile (el mes no avanzaba).
// Cada mes lleva sus propias flechas — apuntan al vecino, así que cada
// sección es su propia ancla (Anchor). El chip colapsado es una barra de
// ancho completo bajo la tira, en flujo normal: visible siempre como toggle
// de plegado, sin flotar ni tapar nada en ningún estado.
//
// El mismo diseño de un mes a la vez ya funciona en mobile sin reglas
// aparte; solo el tamaño de la celda del día crece para un objetivo táctil
// más cómodo.
func (c *CalendarSlider) RenderCSS() *css.Stylesheet {
	return style.For(c).
		// Full width, not Compact: the widget must use its whole column —
		// capped at 24rem it floated centered with dead margins inside a
		// wide aside. The grid tracks (1fr) and the square days do the rest.
		Root(
			style.Center(style.Full),
		).
		// FixedGrid with a real gap (not SpaceNone): the cells grow with
		// their track but the gutters stay one constant Space2 step — the
		// original calendar read as separate cards with even air, never as
		// one fused slab. Same gap header and weeks, so the columns align.
		Part(PartWeekRow,
			style.FixedGrid(7, style.Space2),
		).
		Part(PartStrip,
			style.ScrollRow(style.SpaceNone),
			// The strip shows if and only if expanded — on EVERY viewport,
			// not just mobile: collapsing must work on desktop too (which
			// simply defaults to expanded, as there is room). MotionSlow,
			// not Base: 250ms reads as a cut on a panel this tall, 400ms as
			// a fold — still instant under prefers-reduced-motion.
			style.Animate(style.MotionSlow),
			style.RevealedBy(widget.Current),
		).
		Part(PartMonth,
			style.As(style.Panel),
			style.Pad(style.Space2),
			style.Anchor(),
			style.Width(style.Full),
			style.Stack(style.Space1),
		).
		// The month nav is one static row, identical on every device: prev,
		// label, next centered at the bottom, always in flow, never
		// overlaid. No hover/focus reveal anywhere on purpose: turning the
		// buttons absolute on hover made them vanish under the pointer, and
		// hiding the row on focus cancelled taps on mobile.
		Part(PartMonthNav,
			style.Row(style.Space2),
			style.CenterContent(),
		).
		Part(PartWeekday,
			style.CenterContent(),
			style.FontSize(style.TextXs),
			style.FontWeight(style.WeightBold),
		).
		Part(PartMonthName,
			style.CenterContent(),
			style.FontSize(style.TextSm),
			style.FontWeight(style.WeightBold),
			style.Pad(style.Space1),
		).
		Part(PartDayStack,
			style.Stack(style.SpaceNone),
		).
		// The day fills its grid track and keeps the square shape at any
		// width: MediaBox(AspectSquare) sizes the box off the column, so a
		// 7-column grid never drops below ~24px (today's size, kept as the
		// floor for free) and grows proportionally on wider viewports. No
		// IconBox (a fixed box is what stranded the small cells), no mobile
		// override (aspect covers every breakpoint, touch-sized by geometry).
		// Centering comes with the MediaBox recipe itself.
		Part(PartDay,
			style.MediaBox(style.AspectSquare),
		).
		Part(PartDayNum,
			style.Grow(),
			style.CenterContent(),
			style.FontSize(style.TextSm),
			style.FontWeight(style.WeightBold),
		).
		Part(PartDayUse,
			style.As(style.Inset),
			style.Round(style.RadiusFull),
			style.Meter(style.Space1),
			style.CenterSelf(),
		).
		Part(PartDaySelectable,
			style.Interactive(style.Inset),
		).
		Part(PartDayRed,
			style.Glyph(style.Danger),
		).
		Part(PartDayToday,
			style.As(style.Primary),
			style.Round(style.RadiusSm),
		).
		// The nav arrows are exact 50px squares from closed scales only:
		// FontSize(TextLg) + IconBox(IconLg) = 2.5em of 1.25rem — the same
		// pair that boxes every svg control, with no padding arithmetic and
		// no invented pixels. Transparent (Bare) so the calendar stays
		// light; change --text-xl or the IconSize scale once and every
		// square control follows, no per-component edits.
		Part(PartPrev,
			style.As(style.Bare),
			style.CenterContent(),
			style.FontSize(style.TextLg),
			style.FontWeight(style.WeightBold),
			style.IconBox(style.IconLg),
			style.Interactive(style.Subtle),
			style.Round(style.RadiusSm),
		).
		Part(PartNext,
			style.As(style.Bare),
			style.CenterContent(),
			style.FontSize(style.TextLg),
			style.FontWeight(style.WeightBold),
			style.IconBox(style.IconLg),
			style.Interactive(style.Subtle),
			style.Round(style.RadiusSm),
		).
		When(widget.Selected, PartDay,
			style.As(style.Accent),
			style.Raise(style.Raised),
		).
		// The chip is a full-width bottom bar under the strip, in normal
		// flow — never floating, never overlapping: expanded it reads as
		// the fold control, collapsed the strip above it is gone and the
		// bar stays as the unfold control. No Docked, no anchor, no
		// reserved band needed: the flow lays it out in both states.
		Part(PartCollapsed,
			style.Row(style.Space2),
			style.CenterContent(),
			style.Pad(style.Space2),
			style.As(style.Panel),
			style.Round(style.RadiusSm),
			style.Interactive(style.Subtle),
		).
		Part(PartCollapsedToggle,
			style.Hide(),
		).
		// The calendar glyph wears the same chrome as every other icon cap in
		// the chassis (selectsearch's PartIcon, the footer action buttons): a
		// filled Primary square with a white glyph, not a bare mark. The cap
		// is selectsearch's own recipe — MediaBox(AspectSquare) sized off the
		// same --control-height ControlBox answers to, so all caps are the
		// same square — with a Round of its own (unlike selectsearch's flush
		// cap, this one sits inside the chip's padding, not clipped by it).
		// As() paints the primary gradient and the on-primary (white) text
		// color, which the sprite glyph inherits via currentColor.
		Part(PartCollapsedCap,
			style.As(style.Primary),
			style.MediaBox(style.AspectSquare),
			style.ControlBox(),
			style.KeepSize(),
			style.Round(style.RadiusSm),
			style.CenterContent(),
		).
		// A bare <svg> with no box falls back to 300x150; IconBox pins it.
		// IconLg at TextLg (2.5em of 1.25rem = 50px): the cap is a
		// --control-height square, so the glyph fills it and the whole icon
		// meets the touch-size floor on BOTH axes — IconMd (30px) left the
		// glyph flagged by the mobile audit as a sub-44px hit area.
		// Both are system tokens; no raw pixels, no viewBox padding hacks —
		// svg.go keeps the plain FontAwesome viewBox + currentColor path,
		// the same shape every webtyp/icons glyph package ships.
		Part(PartCollapsedIcon,
			style.FontSize(style.TextLg),
			style.IconBox(style.IconLg),
		).
		Part(PartCollapsedText,
			style.Grow(),
			style.FontSize(style.TextSm),
			style.FontWeight(style.WeightBold),
		).
		Stylesheet()
}
