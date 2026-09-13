//go:build !wasm

package fieldset

import (
	"webtyp.com/css"
	"webtyp.com/widget"
	"webtyp.com/widget/style"
)

// RenderCSS defines the fieldset widget visual contract using the style DSL.
func (f *Fieldset) RenderCSS() *css.Stylesheet {
	return style.For(f).
		// The field is not a card. It is a label chip sitting on an input box:
		// wrapping the pair in a second bordered panel nests a box inside a box
		// and reads as clutter once several fields stack up.
		//
		// The field owns NO vertical rhythm of its own — the gap between
		// consecutive fields is one value, on the form container (PartForm
		// below). What the field reserves is ChipSeat(EdgeTop): exactly half a
		// chip-height of top padding, the space the legend seats into so it
		// centres on the input's top border without poking out of whatever
		// card the form sits in. PadInline is the field's only other padding —
		// the input's inset from the field's own edges.
		Root(
			style.Anchor(),
			style.Stack(style.SpaceNone),
			style.ChipSeat(style.EdgeTop),
			style.PadInline(style.Space2),
			style.KeepSize(),
		).
		// The form: the one place the inter-field rhythm lives. A gap on the
		// container spaces every field the same and — unlike a per-field
		// margin — does not double up at the first and last field against the
		// card's own inset.
		Part(widget.PartForm,
			style.Stack(style.Space3),
		).
		// A legend, not a caption: OnEdge(EdgeTop) seats the chip flush against
		// the field's ChipSeat padding and centres it — pinned to the shared
		// --chip-height token — on the input's top border. SpaceNone block
		// offset: the seat already IS half a chip-height, no extra gap wanted.
		// The targetlist badge straddles its row's bottom line by the same
		// token, so legend and badge align by construction.
		// No ChipBox: a legend is read, not lined up. ChipBox pins every chip to
		// --chip-width and clips the overflow, which is right for a badge in a
		// column of badges (they align) and wrong here — one field per row means
		// there is no column to align with, so a short label like "id" padded a
		// fixed 112px reads as a mislabelled box, and a long one is silently cut
		// mid-word. Shrink-wrapped to its text, the chip says exactly its own
		// name at any length.
		// No ChipBox: a legend is read, not lined up. ChipBox pins every chip to
		// --chip-width and clips the overflow, which is right for a badge in a
		// column of badges (they align) and wrong here — one field per row means
		// there is no column to align with, so a short label like "id" padded a
		// fixed 112px reads as a mislabelled box, and a long one is silently cut
		// mid-word. Shrink-wrapped to its text, the chip says exactly its own
		// name at any length.
		//
		// Capitalize: the text arrives as a model's field name (id, name, ip),
		// lower-cased because that is what the struct tag says. Casing it here
		// keeps that decision in the skin instead of asking every model to
		// pre-format a display string.
		Part(widget.PartLabel,
			style.OnEdge(style.EdgeTop, style.SideStart, style.SpaceNone, style.Space4),
			style.As(style.Primary),
			style.Round(style.RadiusSm),
			style.FontSize(style.TextXs),
			style.FontWeight(style.WeightBold),
			style.Raise(style.Raised),
			style.Capitalize(),
			style.StartContent(),
			style.PadInline(style.Space2),
		).
		// Red text growing leftward from the field's trailing edge — not a filled
		// bar across the whole field, and not riding the border either: OnEdge
		// straddled the input's top line and the message came out cut by it.
		// It sits inside the box with equal air above and to the right.
		// Space4, not Space2: an absolute box is laid out against the Root's
		// PADDING box, so the first Space2 only buys back the Root's own Pad and
		// leaves the text flush with the input's border. Space4 spends that Space2
		// and puts a real Space2 between the message and both edges.
		Part(widget.PartError,
			style.Docked(style.Parent, style.EdgeTop, style.SideEnd, style.Space4),
			style.Glyph(style.Danger),
			style.FontSize(style.TextXs),
		).
		// Space4, not Space2: the legend rides the input's top border and hangs
		// half its height inside the box, so the value needs room to clear it.
		// At the chip's 20px height the hang is 10px and Space4's 16px still
		// clears it.
		// Panel, not Page: the legend chip is centred ON the input's top border
		// line, and a borderless Page box gave it no line to sit on — the chip
		// then read as floating a hair high, anchored to an edge the eye could
		// not see. Panel's hairline ColorOutline frame is the same one the
		// search bar's own input carries, so every text control in the app now
		// shares one framed shape; the focus ring (amber, inset — see
		// css/css.reset.go) lands just inside that frame instead of fighting it.
		Part(widget.PartInput,
			style.As(style.Panel),
			style.Round(style.RadiusMd),
			style.Pad(style.Space4),
			style.ControlBox(),
		).
		// Bottom-right corner of the field, not top-right: PartError already
		// owns the top-right corner (the two would overlap the moment a
		// masked field also has an error to show). Docked has no
		// vertical-center primitive today — a real gap in the DSL, not
		// solved here — so bottom-right is the closest fit that never
		// collides with the label chip (top) or the error message (top,
		// trailing). ControlBox matches PartInput/PartSubmit's own tap
		// target instead of a bespoke size. Inactive by default (a quiet
		// affordance, not a second call to action next to the field it
		// sits in) — When(Selected, …) below brightens it once revealed.
		Part(widget.PartReveal,
			style.Docked(style.Parent, style.EdgeBottom, style.SideEnd, style.Space2),
			style.Glyph(style.Inactive),
			style.ControlBox(),
		).
		Part(widget.PartRadioGroup,
			style.Row(style.Space3),
			style.Pad(style.Space1),
		).
		// Filled and full-width, matching crudview's own primary action
		// button exactly (As(Primary), Pad(Space3), ControlBox) -- a form's
		// submit is the same weight of commitment as crudview's own save
		// action, and a bare unstyled <button> next to a skinned field stack
		// was the one piece of every form (login included) that still read
		// as unfinished.
		Part(widget.PartSubmit,
			style.Button(style.Primary),
			style.Width(style.Full),
			style.CenterContent(),
		).
		// Locked repaints the INPUT, not the field root — historically the
		// wrapper got the treatment and the resulting ring crossed the legend
		// chip straddling the top border line. Nothing about a locked field
		// needs a second box around the pair; the control itself is what is
		// locked, so the control is what changes. (Secondary drops the base
		// Panel's hairline frame and repaints the fill — no border, no ring;
		// the only box-shadow in this sheet stays the label chip's Raise
		// elevation.)
		//
		// This gate is PER FIELD, not per form. dom writes data-locked on the
		// wrapper whenever form's isDisabledOrLocked() says so, and that check
		// ORs the whole-form lock (SetLocked — no longer called by crudview,
		// but still public API) with the field's OWN
		// fc.Input.IsDisabled(), a per-field property declared on the model.
		// Without this rule, a genuinely disabled field would be
		// indistinguishable from an editable one — which is why the rule
		// survives the form-lock's retirement.
		//
		// Secondary is Panel's fill with no border: the control stops reading
		// as a writing area and settles back into the card, which is what a
		// value you are only reading should do. The text stays ColorOnSurface
		// at full contrast — a read-only value is still there to BE read, so
		// this deliberately does not reach for Inactive's muted text.
		// WhenWithin, not When: dom writes data-locked on the field wrapper
		// that owns the gate, so When(Locked, PartInput) would emit
		// `.NAME__input[data-locked="true"]` (NAME = widget.NameField) and match
		// nothing.
		// Round(RadiusMd) is not redundant: a surface with no explicit radius
		// falls back to its own default, and Secondary's is RadiusSm, so
		// locking silently shrank the control's corners from the 8px the base
		// rule gives it to 4px. Restating the base radius keeps the box the
		// same shape in both states — only its fill is supposed to change.
		WhenWithin(widget.Locked, "", widget.PartInput,
			style.As(style.Secondary),
			style.Round(style.RadiusMd),
		).
		// No filled state on the field itself: the red message in the corner is
		// the signal. Painting the whole box danger buried the value.
		When(widget.Invalid, widget.PartInput,
			style.Glyph(style.Danger),
		).
		// dom writes data-selected on the reveal button itself when the
		// value is showing (see form's wireMaskToggle) — When targets the
		// same element the state is written on, unlike WhenWithin above.
		When(widget.Selected, widget.PartReveal,
			style.Glyph(style.Primary),
		).
		Stylesheet()
}
