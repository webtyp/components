//go:build !wasm

package selectsearch

import (
	"webtyp.com/components/listgap"
	"webtyp.com/css"
	"webtyp.com/widget"
	"webtyp.com/widget/style"
)

// RenderCSS defines the selectsearch visual contract using the style DSL.
func (c *SelectSearch) RenderCSS() *css.Stylesheet {
	return c.sheet().Stylesheet()
}

// sheet is split out from RenderCSS so tests can call Validate() on the
// *style.Sheet directly, without the *css.Stylesheet conversion — the same
// split targetlist/css.go uses (see targetlist_test.go's TestSheetValidates).
func (c *SelectSearch) sheet() *style.Sheet {
	s := style.For(c).
		// Anchor: what lets PartDropdown's Flyout (added below, tablet/desktop
		// only) hang from THIS box instead of some positioned ancestor further
		// up the tree. Mirrors usermenu's Root(Anchor()) exactly — same
		// primitive, same reason ("a corner pin needs a positioned ancestor to
		// hang from").
		Root(
			style.Anchor(),
			style.Stack(style.Space1),
			style.As(style.Panel),
			style.Round(style.RadiusMd),
		)
	// The same list container targetlist and targetdate assemble — one lego
	// piece, so the dropdown's scroll region, its block gutter and its lateral
	// inset are the ones every other list in the app uses.
	//
	// The row RHYTHM is the one value taken back, and only that one: listgap's
	// gap is documented as sized to clear PartBadge's straddle — a badge that
	// hangs half a chip-height BELOW its row's bottom edge, so the rows have to
	// stand apart far enough not to be overlapped (16px desktop / 24px mobile).
	// selectsearch has no such badge: PartDesc is an inline chip that lives
	// INSIDE the row's box. Inheriting a gap sized for a constraint this list
	// does not have was invisible while each option carried its own card and
	// the card filled the space; with PartOption on the sheet's own ground
	// (see Interactive(Bare) below) the rows read as adrift.
	//
	// SpaceNone, not a smaller gap: what separates one option from the next is
	// a hairline (DividerBelow on PartOption), and a hairline only reads as a
	// SEPARATOR when the rows it divides are flush. Leave a gap and the same
	// line reads as an underline drawn under each row — the rule stops
	// belonging to the pair and starts belonging to the row above it.
	// Declared after the spread on purpose: Part()/On() merge, so the later
	// Stack() is the one that lands.
	//
	// DividerBetween lives on the CONTAINER, not on the row: it emits on every
	// child that has one before it, so four options get three rules and neither
	// end dangles. DividerBelow() on PartOption would have left a hairline
	// under the last row with nothing after it to separate — a list that reads
	// as cut off rather than as ended.
	listgap.Apply(s, PartOptions)
	s.Part(PartOptions, style.Stack(style.SpaceNone), style.DividerBetween())
	s.On(css.Mobile, PartOptions, listgap.MobileOpts()...)
	s.On(css.Mobile, PartOptions, style.Stack(style.SpaceNone))
	return s.
		// The checkbox drives PartDropdown's open state via the label's `for`
		// attribute (see selectsearch.go) — it is the CSS-only toggle
		// mechanism, never meant to be seen. A label still activates a
		// display:none checkbox natively, so Hide() loses nothing.
		Part(PartToggle,
			style.Hide(),
		).
		// One shape at every breakpoint: a floating sheet hanging off the Root
		// anchor. This USED to be an in-flow accordion on a phone, on the
		// grounds that "there is no room for an anchored overlay on a narrow
		// viewport" — but the thing that had no room was an UNBOUNDED panel.
		// Measured at 320x568 with 20 patients: the in-flow dropdown grew to
		// 1856px, overflowing the viewport by 1346px, and what moved under the
		// user's thumb was the whole aside. Capped() removes that reason, so
		// the breakpoint split goes with it: one code path, one shape, and the
		// phone gets the SAME "this is temporary and on top" reading the
		// desktop always had.
		//
		// As(Page), not As(Panel): the pane this hangs over is already painted
		// ColorSurface (measured: rp__aside and the old dropdown were both
		// #F2F2F7 — the overlay was literally the same colour as its own
		// container, separated by 0.8px of hairline). A sheet has to be a
		// DIFFERENT ground from what it covers or it is not a sheet. Page
		// brings no radius of its own (Page.defaultRadius() is RadiusNone), so
		// Round is explicit here.
		//
		// Raise(Popover) at every breakpoint too — and note this is elevation,
		// not position. The old code coupled the two and dropped both on
		// mobile; but a shadow costs no layout space, and a phone is precisely
		// where there is no other depth cue to fall back on.
		//
		// Capped + the inner Scroll() are a PAIR: neither works alone. See
		// Capped's own doc for why every declaration Scroll() emits is inert
		// until an ancestor has a definite block size.
		//
		// Pad(Space1): the sheet is rounded and clipping (HideOverflow), so a
		// child flush against its top edge loses its corners — and with them
		// the inset focus ring, which is what made the search box look sawn
		// off. The padding is what keeps every child inside the rounded area.
		Part(PartDropdown,
			style.Stack(style.Space1),
			style.As(style.Page),
			style.Round(style.RadiusMd),
			style.Pad(style.Space1),
			style.HideOverflow(),
			style.Capped(style.ExtentMost),
			style.Raise(style.Popover),
			style.Flyout(style.SideStart),
			style.Width(style.Full),
		).
		// The scrim. This is the cue that answers "which list am I using?" —
		// the other list stops looking actionable instead of merely looking
		// different. usermenu's PartBackdrop is the same recipe (Backdrop +
		// RevealedBy, declared BEFORE the panel so the two tie at the dropdown
		// layer and DOM order puts the panel on top); the difference is that
		// this one is not gated on a breakpoint, because the confusion it
		// fixes is worst on a phone.
		Part(PartBackdrop,
			style.Backdrop(style.Viewport),
			style.Veil(),
		).
		// SpaceNone, not Space2: the gap between the icon and the content comes
		// from PartHeaderBody's own left padding below, same as searchbar's
		// root uses a zero gap and lets PartInput's own Pad(Space2) make the
		// space — a Row gap here would ADD to that, pushing the icon away
		// from the header's own edge instead of leaving it flush.
		// Round+HideOverflow (not on Root, where the floating PartDropdown
		// lives — clipping there would cut the dropdown off) is what makes
		// PartIcon's own square corners read as the header's rounded ones:
		// the icon carries no radius of its own, the header clips it to
		// match.
		// ControlBox+KeepSize on BOTH the bar and its cap, exactly as
		// searchbar/css.go declares them: Row() carries flex-wrap: wrap, so a
		// narrow viewport wrapped the text onto its own line under the square
		// and the header stopped being a bar. KeepSize is what forbids the
		// wrap; ControlBox is the FLOOR height (min-height) — with a picked
		// patient the two-line body makes the header taller than one control,
		// and that is fine.
		Part(PartHeader,
			style.Row(style.SpaceNone),
			style.Round(style.RadiusMd),
			style.HideOverflow(),
			style.ControlBox(),
			style.KeepSize(),
		).
		// The header's content area, beside the flush icon cap. It carries the
		// inline padding (the cap must not — see PartHeader) so the name column
		// and the trailing time chip both clear the header's rounded clip, and
		// it is a Row so those two lay out exactly as they do inside an option:
		// PartText grows, PartDesc is pushed to the trailing edge. Grow() so
		// the body itself takes all the width the cap leaves.
		//
		// PadInline, not Pad: block padding here would stack on top of the two
		// text lines and push the header past its --control-height floor,
		// unflushing the cap. The block breathing room instead comes from the
		// header's min-height being taller than the text — the lines centre in
		// the leftover space.
		Part(PartHeaderBody,
			style.Grow(),
			style.Row(style.Space2),
			style.PadInline(style.Space2),
		).
		// The empty-state line: shown only until a choice is made. Subtle so it
		// reads as a prompt, not as content — the same muted treatment
		// PartSublabel wears. Once a patient is picked this is replaced by the
		// PartText column, which is bold.
		Part(PartPlaceholder,
			style.Glyph(style.Subtle),
			style.Grow(),
		).
		// The filled square cap around the arrow, FIRST child of the header
		// (see selectsearch.go) — searchbar's own PartIcon recipe
		// (MediaBox(AspectSquare) sized off the same --control-height
		// ControlBox answers to, so the cap is a true square), As(Primary)
		// so it reads as part of the same chrome the rest of the chassis
		// uses instead of a bare mark floating in the header. No Round of
		// its own — PartHeader's clip above is what rounds its visible
		// corners. KeepSize, as on the header: a square cannot wrap into a
		// sliver when the row next to it runs out of space.
		//
		// The cap stays a --control-height square even with a picked patient:
		// PartHeaderBody below pads on the inline axis only, so its two text
		// lines (~48px) sit inside the header's --control-height floor without
		// pushing it taller — the header stays one control tall and the
		// centred cap fills it edge to edge.
		Part(PartIcon,
			style.As(style.Primary),
			style.IconCap(),
		).
		// PartGlyph survives IconCap for ONE reason: the GLYPH turns, not
		// PartIcon — rotating the cap would spin the whole filled square. It
		// carries no IconBox any more; the cap sizes its <svg> (half the cap)
		// and that is no longer this component's decision. TurnNone is the
		// resting rule Animate() transitions from — without a base value
		// there is no start state to move off.
		Part(PartGlyph,
			style.Rotate(style.TurnNone),
			style.Animate(style.MotionBase),
		).
		// A radius of its own so the box reads as a control, not as a slab
		// filling the panel's width. ControlBox answers to the same
		// --control-height token the header uses, so the search field and the
		// cap stay the same height.
		//
		// KeepSize is what pins it once PartDropdown is Capped(): the sheet is
		// a flex column with a ceiling, and every child of a bounded flex
		// container is shrinkable by default. Without it the search box is the
		// first thing the list squeezes as options pile up, and the field the
		// user is typing into collapses while they type.
		Part(PartSearch,
			style.Pad(style.Space2),
			style.As(style.Inset),
			style.Round(style.RadiusSm),
			style.ControlBox(),
			style.KeepSize(),
		).
		// Interactive(Bare), not Interactive(Page): a row on a floating sheet
		// has no card of its own. This deliberately walks back "a dropdown
		// option and a list row are the same object, they must not be two
		// different shapes" — that parity is what made the two lists
		// indistinguishable (measured: both #FFFFFF, both 8px radius, both
		// 12px padding, both amber when selected, stacked one above the
		// other). The parity that MATTERS — box height, rhythm, the amber
		// "selected" statement — is kept below; only the resting card is
		// dropped, because on a sheet the sheet IS the card. Round stays: it
		// shapes the amber fill the Cue rules paint on hover and selection.
		//
		// Round(RadiusNone) goes with the container's DividerBetween: a rounded
		// fill between two hairlines leaves a wedge of sheet showing at each
		// corner, so the highlight stops meeting the rule that is supposed to
		// bound it. Square corners let the amber run the full width between
		// hairlines, which is what a menu highlight has always looked like.
		// (platformd's mobile nav-link pairs the same two for the same reason.)
		//
		// Animate(MotionFast): Interactive() drives four background/text swaps
		// on this row — transparent → wash on hover/focus, → full Accent on
		// select/press — and with no transition each is an instant repaint.
		// One row snapping from clear to solid amber under the pointer is the
		// "abrupt, strange" jump; 150ms (MotionFast, the scale step named for
		// "immediate highlight: hover, focus") is enough to read as a fade
		// without feeling slow. `all` also covers the colour flip, which is
		// the point. The DSL silences it under prefers-reduced-motion on its
		// own — hasMotion is what emitReducedMotion scans for.
		Part(PartOption,
			style.Row(style.Space2),
			style.KeepSize(),
			style.ControlBox(),
			style.As(style.Bare),
			style.Interactive(style.Bare),
			style.Pad(style.Space3),
			style.Round(style.RadiusNone),
			style.Animate(style.MotionFast),
		).
		// The Label/Sublabel column: Grow() takes the row's free width so
		// PartDesc's badge lands at the trailing edge instead of hugging
		// straight after the name — same reasoning as targetlist's own
		// PartLabel.
		Part(PartText,
			style.Stack(style.SpaceNone),
			style.Grow(),
		).
		Part(PartLabel,
			style.FontWeight(style.WeightBold),
		).
		// Muted and small — a second line under the name, not competing with
		// it for attention. Glyph, not As(Inset): a tinted color reads as
		// secondary text, a filled box reads as its own control — a plain
		// two-line name+id has no control to be.
		Part(PartSublabel,
			style.Glyph(style.Subtle),
			style.FontSize(style.TextXs),
		).
		// The trailing badge. PadInline is not optional decoration here: the
		// chip is a FILLED box, and text flush against a filled edge reads as
		// clipping, not as density — which is exactly what "09:00" looked
		// like with a background and no inset. PadInline rather than Pad
		// because the vertical size is contracted against the row, not chosen
		// here; see its doc in widget/style, written for this shape of chip.
		//
		// Deliberately NOT the ChipBox() recipe targetlist/targetdate/usermenu
		// use for their badges: ChipBox pins --chip-width (7rem), which is the
		// right call for a truncating label of unknown length and the wrong
		// one for a five-character time — it would spend a third of a 375px
		// row on "09:00". CenterContent still applies, so should this slot
		// ever carry a shorter string it stays centred in its own box.
		Part(PartDesc,
			style.As(style.Inset),
			style.FontSize(style.TextXs),
			style.PadInline(style.Space2),
			style.CenterContent(),
			style.KeepSize(),
		).
		// The chevron IS the open state. WhenWithin from the root, because
		// data-open lives on the root (see selectsearch.go's BindState) and the
		// glyph is two levels down inside the header.
		WhenWithin(widget.Open, "", PartGlyph,
			style.Rotate(style.TurnHalf),
		).
		// The cap goes amber while the list is open, tying the open sheet back
		// to the control that opened it — otherwise the trigger looks identical
		// open or closed and the sheet reads as free-floating chrome that
		// belongs to nothing. AccentInverse is the surface written for exactly
		// this case; its own doc in widget/style/surface.go describes it as
		// "the 'add' button gone amber while its panel is open". Being a
		// derived surface it also emits background-image: none, which is what
		// clears the Primary family gradient underneath — without that the
		// blue gradient paints straight over the amber.
		WhenWithin(widget.Open, "", PartIcon,
			style.As(style.AccentInverse),
		).
		// Amber is the chassis' one "where I am" statement — the rail's current
		// nav item and a selected list row wear it too. AccentWash on hover and
		// focus reads as "on the way to selected"; a grey mix reads as chrome
		// with no relation to what clicking does. Focus repeats Hover on
		// purpose: a keyboard user gets no :hover and must not be stranded on a
		// different colour.
		//
		// Round(RadiusNone) is repeated on every one of them, and it is not
		// noise: a state rule with no Round of its own falls back to the
		// SURFACE's default radius (Accent and AccentWash both resolve to
		// RadiusSm), and @layer states outranks @layer widgets — so the base
		// rule's square corners were being restored to 4px the moment a row
		// was hovered or selected, leaving a wedge of sheet showing at each
		// corner of the fill the hairlines were supposed to bound.
		When(widget.Selected, PartOption, style.As(style.Accent), style.Round(style.RadiusNone)).
		Cue(widget.Hover, PartOption, style.As(style.AccentWash), style.Round(style.RadiusNone)).
		Cue(widget.Focus, PartOption, style.As(style.AccentWash), style.Round(style.RadiusNone)).
		Cue(widget.Press, PartOption, style.As(style.Accent), style.Round(style.RadiusNone))
	// No per-breakpoint escalation for PartDropdown any more: the sheet, its
	// elevation and its cap are declared once in the base rule above and apply
	// everywhere. The old On(Tablet)/On(Desktop) pair existed only to withhold
	// Flyout+Raise from phones; Capped() removed the reason, and one shape at
	// every width is one shape to reason about.
}
