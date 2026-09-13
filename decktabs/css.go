//go:build !wasm

package decktabs

import (
	"webtyp.com/css"
	"webtyp.com/widget"
	"webtyp.com/widget/style"
)

// RenderCSS defines the stylesheet for tabs.
func (t *DeckTabs) RenderCSS() *css.Stylesheet {
	return style.For(t).
		// Strip above deck, deck taking the rest: the root is a column whose
		// second child is the one that grows.
		Root(
			style.Stack(style.SpaceNone),
			style.Fill(),
		).
		// KeepSize so the strip never gives up its height to the deck, and a
		// hairline so the tabs read as sitting on top of the panel rather than
		// floating above unrelated content.
		Part(PartList,
			style.Row(style.Space2),
			style.KeepSize(),
			style.PadInline(style.Space2),
			style.DividerBelow(),
		).
		// Button(), not a hand-composed Interactive+ControlBox: a tab is
		// pressed, and Button is the one recipe for that — it carries the
		// shared control height, the inline padding, and a box that neither
		// stretches to the strip nor shrinks under pressure. Subtle is the
		// resting family, so an unselected tab stays quiet.
		Part(PartTab,
			style.Button(style.Subtle),
			style.Row(style.Space2),
			style.CenterContent(),
		).
		// The selected tab is the one filled chip in the strip — the same
		// "current" language layout/platformd uses for the route you are on.
		// AccentInverse keeps the icon legible against the committed fill
		// through currentColor. No hover rule here: Button already derives
		// hover, focus and press from its own family, and a second one would
		// be redundant composition.
		When(widget.Current, PartTab,
			style.As(style.AccentInverse),
		).
		// Layers, not a strip. SlideDeck stacks the panels and reveals the one
		// carrying widget.Current, sliding it in from the inline start.
		//
		// It is deliberately NOT a scroller: a horizontal scroller here would
		// chain with the horizontal scroll a panel's own content may have, and
		// a swipe inside the content would end up changing tab on its own.
		Part(PartDeck,
			style.SlideDeck(style.MotionBase),
			style.Fill(),
		).
		// Scroll(), not Fill(): a panel taller than its layer has to scroll
		// inside that layer. Scroll() is Fill() plus overflow-y.
		//
		// No Anchor(): SlideDeck already positions each layer absolutely, which
		// makes the panel the containing block for its own content. Anchor()
		// would emit position:relative in @layer widgets, beating the
		// position:absolute the flow emits in @layer primitives, and the layers
		// would fall back into normal flow stacked one below the other.
		Part(PartPanel,
			style.Stack(style.SpaceNone),
			style.Scroll(),
		).
		Part(PartIcon,
			style.IconBox(style.IconSm),
			style.KeepSize(),
		).
		Part(PartLabel,
			style.FontSize(style.TextSm),
		).
		Stylesheet()
}
