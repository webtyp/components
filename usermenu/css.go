//go:build !wasm

package usermenu

import (
	"webtyp.com/css"
	"webtyp.com/widget"
	"webtyp.com/widget/style"
)

// RenderCSS defines the usermenu visual contract using the style DSL.
func (m *UserMenu) RenderCSS() *css.Stylesheet {
	return style.For(m).
		// Anchor, not a flow: the panel hangs off this element, and a corner
		// pin needs a positioned ancestor to hang from.
		Root(
			style.Anchor(),
		).
		// The resting state: avatar and name, nothing else. Whatever a shell
		// puts here has to have a bounded width, and a name does.
		Part(PartTrigger,
			style.Row(style.Space2),
			style.Button(style.Subtle),
			// A pill, not a 4px-cornered box: Button derives only RadiusSm from
			// Subtle, so the explicit radius stays.
			style.Round(style.RadiusFull),
			style.Pad(style.Space1),
		).
		// IconMd, not IconLg: the trigger lives in the platform header, and a
		// 40px avatar inflated the bar past 65px. At 24px the header settles
		// near 40px. The brand mark mirrors this box — change both or neither.
		Part(PartAvatar,
			style.IconBox(style.IconMd),
			style.Round(style.RadiusFull),
			style.HideOverflow(),
		).
		Part(PartName,
			style.FontSize(style.TextBase),
			style.FontWeight(style.WeightBold),
			style.KeepSize(),
		).
		// In flow by default, which is what a phone wants: the menu lives inside
		// a drawer there, and a flyout sized to its content is wider than the
		// drawer — it would hang off the side as a second floating panel over
		// the content. Opening it should expand the drawer entry instead.
		Part(PartPanel,
			style.Stack(style.Space2),
			style.As(style.Panel),
			style.Round(style.RadiusMd),
			style.Pad(style.Space2),
			style.Width(style.Full),
		).
		Part(PartRoles,
			style.Row(style.Space1),
		).
		// The same chip box the rest of the system uses, so a role beside a
		// field legend or a list badge reads as the same kind of object.
		Part(PartRole,
			style.As(style.Inset),
			style.Round(style.RadiusSm),
			style.FontSize(style.TextXs),
			style.ChipBox(),
			style.CenterContent(),
		).
		Part(PartActions,
			style.Row(style.Space1),
		).
		// Keeping its size is a HEADER concern: there the menu shares a row with
		// a message slot that will happily squeeze it. On a phone the menu sits
		// in a drawer narrower than its own content, and refusing to shrink
		// there makes it size to its widest role chip and hang out over the
		// content as a panel of its own.
		On(css.Tablet, "", style.KeepSize()).
		On(css.Desktop, "", style.KeepSize()).
		// From a tablet up the menu sits in a header row, where an inline panel
		// would push that row open and shove the content beneath it down every
		// time someone looked at who they were logged in as. There it hangs —
		// and a hanging panel needs a backdrop to catch the click that dismisses
		// it. Both are one mechanism, so both are gated on the same viewports:
		// on a phone the menu is an accordion inside the drawer, the drawer's
		// own overlay already dismisses it, and a viewport-sized backdrop there
		// would sit ON TOP of an in-flow panel and eat every tap meant for the
		// controls inside it.
		On(css.Tablet, PartPanel,
			style.Width(style.Content),
			style.Raise(style.Floating),
			style.Flyout(style.SideEnd),
		).
		On(css.Tablet, PartBackdrop,
			style.Backdrop(style.Viewport),
			style.RevealedBy(widget.Open),
		).
		On(css.Desktop, PartPanel,
			style.Width(style.Content),
			style.Raise(style.Floating),
			style.Flyout(style.SideEnd),
		).
		On(css.Desktop, PartBackdrop,
			style.Backdrop(style.Viewport),
			style.RevealedBy(widget.Open),
		).
		Stylesheet()
}
