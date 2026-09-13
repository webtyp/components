//go:build !wasm

package calendarslider

import (
	"webtyp.com/css"
	"webtyp.com/widget"
	"webtyp.com/widget/style"
)

// RenderCSS describe el aspecto del calendario con los tokens del tema: un
// mes a la vez, centrado (Center), con los meses vecinos fuera de vista en
// una tira de scroll-snap (ScrollRow) — cada mes ocupa el 100% de la tira,
// así que solo uno es visible; el hoy se rellena con el color primario y las
// flechas ‹ › son una fila compacta centrada junto al nombre del mes
// (PartMonthNav), idéntica en desktop y táctil. Es estática a propósito:
// sin reveal en hover/foco — volver las flechas absolutas las hacía
// desaparecer bajo el puntero, y ocultar la fila con foco cancelaba los
// toques en mobile (el mes no avanzaba).
// Cada mes lleva sus propias flechas — apuntan al vecino, así que cada
// sección es su propia ancla (Anchor). El campo colapsado es la primera
// fila del componente, en flujo normal: el disparador arriba y el panel del
// mes colgando de él, como cualquier date picker.
//
// El mismo diseño de un mes a la vez ya funciona en mobile sin reglas
// aparte: la celda sale de la escala de controles, no del contenedor, así
// que no hace falta ninguna regla de dispositivo.
func (c *CalendarSlider) RenderCSS() *css.Stylesheet {
	return style.For(c).
		// Compact (min(100%, 24rem)), NOT Full: a day cell is a CONTROL, not
		// content. It holds a number and a 4px meter — nothing inside it
		// benefits from more room, so it must not grow. An agenda view
		// (Google Calendar, FullCalendar) grows its cells because each holds
		// events; a picker (native input[type=date], Material's 328px, Ant)
		// keeps one fixed box at every viewport. This is a picker.
		//
		// Full is what made desktop unusable: it emits max-width: 100%, so
		// PartDay's aspect-ratio squared off an unbounded track — 164px cells
		// and ~1000px of card height on a 1214px page, with the day number
		// stranded at TextSm (8% of the cell). The earlier note here blamed
		// "dead margins inside a wide aside", but rightpanel's
		// Split(SplitTwoThirds) gives the aside a THIRD of the frame (~400px
		// at 1200px), which 24rem fills almost exactly. The dead margins came
		// from web/client.go rendering the demo in a bare full-width div.
		//
		// The arithmetic closes itself: (384 - 16 pad - 48 gaps) / 7 = 45.7px
		// per cell, above the 44px touch floor, and a 14px number is 30% of
		// the cell — the standard picker proportion. A 400px phone resolves
		// min(100%, 24rem) to the same box: one value, zero breakpoints.
		Root(
			style.Center(style.Compact),
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
		// label, next centered at the TOP of the card, always in flow, never
		// overlaid. At the top because the label names the grid below it —
		// you have to know which month you are reading before you read the
		// days; the old bottom placement only mirrored the legacy widget,
		// and legacy fidelity is not a UX argument. No hover/focus reveal
		// anywhere on purpose: turning the buttons absolute on hover made
		// them vanish under the pointer, and hiding the row on focus
		// cancelled taps on mobile.
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
		// A selectable day renders this <button> in place of PartDayStack, so
		// the day is reachable by keyboard and Enter/Space select it for free
		// — the <li role=gridcell> carried the click handler with no tabindex,
		// which made every day unreachable by Tab while the widget still
		// declared role=grid. A button inside a gridcell is the standard
		// pattern (Material does the same).
		//
		// Bare strips the UA button chrome (the same recipe PartPrev/PartNext
		// use), and Fill()+Width(Full) make the button the WHOLE cell: the hit
		// area has to stay the 46px square it was when the <li> was the
		// click target, not shrink to the glyph. Fill's height:100% resolves
		// because PartDay's aspect-ratio gives the cell a definite height.
		Part(PartDayButton,
			style.As(style.Bare),
			style.Stack(style.SpaceNone),
			style.Fill(),
			style.Width(style.Full),
		).
		// The day fills its grid track and keeps the square shape:
		// MediaBox(AspectSquare) sizes the box off the column. The aspect was
		// never the bug — the unbounded Root was. With Root capped at 24rem
		// the track is bounded too, so the square lands at ~46px on every
		// viewport: above the 44px touch floor, and no mobile override needed
		// (aspect covers every breakpoint, touch-sized by geometry). No
		// IconBox — a fixed box is what stranded the small cells.
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
		// PartDayUse carries the SHAPE of the meter; the three parts below
		// carry its colour. Meter() emits only a height and
		// width: var(--meter-fill), so there is no track behind the bar to
		// judge the fraction against — painted Inset on an Inset cell, 60%
		// and 80% were two indistinguishable grey dashes. Occupancy is the
		// point of a booking calendar, so the level has to survive a glance.
		//
		// One hue family with a washed middle step reads as a ramp — free,
		// filling, full — using only surfaces the theme already ships, so no
		// invented values enter the stylesheet.
		// PartDayUse carries size and placement ONLY — no As() of its own.
		// Every meter always renders with one of the three level parts below,
		// and two stacked surfaces on one element fight: the level's As()
		// re-declares border-radius, so an Inset base here lost the pill
		// shape to the level's RadiusSm, and Inset's 1px outline ate half of
		// a 4px-tall bar. One element, one surface.
		Part(PartDayUse,
			style.Meter(style.Space1),
			style.CenterSelf(),
		).
		// El trazo de disponibilidad: color primario, un solo estado. No entra
		// en la rampa de ocupación — ver useLevelClass.
		Part(PartDayUseOn,
			style.As(style.Primary),
			style.Round(style.RadiusFull),
		).
		Part(PartDayUseLow,
			style.As(style.Success),
			style.Round(style.RadiusFull),
		).
		Part(PartDayUseMid,
			style.As(style.DangerWash),
			style.Round(style.RadiusFull),
		).
		Part(PartDayUseHigh,
			style.As(style.Danger),
			style.Round(style.RadiusFull),
		).
		Part(PartDaySelectable,
			style.As(style.Inset),
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
			style.Interactive(style.Panel),
			style.Round(style.RadiusSm),
		).
		Part(PartNext,
			style.As(style.Bare),
			style.CenterContent(),
			style.FontSize(style.TextLg),
			style.FontWeight(style.WeightBold),
			style.IconBox(style.IconLg),
			style.Interactive(style.Panel),
			style.Round(style.RadiusSm),
		).
		When(widget.Selected, PartDay,
			style.As(style.Accent),
			style.Raise(style.Raised),
		).
		// The field is the component's FIRST row, above the strip, in normal
		// flow — never floating, never overlapping: expanded it reads as the
		// fold control, collapsed the strip below it is gone and the field
		// stays as the unfold control. Above and not below because it is the
		// trigger: every date picker puts the field first and hangs the month
		// panel off it, and this one used to sit under the panel it opened.
		// No Docked, no anchor, no reserved band: the flow lays it out in
		// both states.
		//
		// No Pad(): PartCollapsedCap's ControlBox already sets the row to
		// --control-height (50px), the ecosystem's field height. The padding
		// stacked on top of that and made the row 66px — the dead band the
		// mobile layout was paying for in BOTH states.
		Part(PartCollapsed,
			style.Row(style.Space2),
			style.CenterContent(),
			style.As(style.Panel),
			style.Round(style.RadiusSm),
			style.Interactive(style.Panel),
		).
		Part(PartCollapsedToggle,
			style.Hide(),
		).
		// The calendar glyph wears the same chrome as every other icon cap in
		// the chassis (searchbar's PartIcon, selectsearch's): a filled Primary
		// square with a white glyph, not a bare mark. IconCap is that cap —
		// one recipe instead of the four options this used to spell out, and
		// it sizes the <svg> inside it too (half the cap), which is the part
		// the four never covered. As() paints the primary gradient and the
		// on-primary (white) text colour the glyph inherits via currentColor.
		//
		// This component is why IconCap exists: composed by hand, the glyph
		// was sized IconLg+TextLg — 50px inside a 50px cap — and a SOLID
		// glyph (this calendar's path spans its whole 448x512 viewBox, unlike
		// a thin one such as icons/plus) then filled the square edge to edge.
		Part(PartCollapsedCap,
			style.As(style.Primary),
			style.IconCap(),
		).
		Part(PartCollapsedText,
			style.Grow(),
			style.FontSize(style.TextSm),
			style.FontWeight(style.WeightBold),
		).
		Stylesheet()
}
