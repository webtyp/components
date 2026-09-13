package decktabs

import (
	. "webtyp.com/dom"
	. "webtyp.com/html"
	"webtyp.com/svg"
	"webtyp.com/widget"
)

// NameDeckTabs is the widget name for decktabs.
const NameDeckTabs = widget.Name("decktabs")

const (
	PartList  = widget.Part("list")
	PartTab   = widget.Part("tab")
	PartIcon  = widget.Part("icon")
	PartLabel = widget.Part("label")
	PartDeck  = widget.Part("deck")
	PartPanel = widget.Part("panel")
)

var (
	clsRoot  = NameDeckTabs.Root()
	clsList  = NameDeckTabs.Class(PartList)
	clsTab   = NameDeckTabs.Class(PartTab)
	clsIcon  = NameDeckTabs.Class(PartIcon)
	clsLabel = NameDeckTabs.Class(PartLabel)
	clsDeck  = NameDeckTabs.Class(PartDeck)
	clsPanel = NameDeckTabs.Class(PartPanel)
)

// Item is one tab and the panel it reveals.
type Item struct {
	// ID identifies the tab. It is what Active carries and what the ARIA
	// attributes link by, so it must be unique within one DeckTabs and stable
	// across renders.
	ID string
	// Label is the text on the tab. Supplied by the consumer already in the
	// reader's language — this component translates nothing.
	Label string
	// Icon is optional; the empty Icon renders none.
	Icon svg.Icon
	// Panel is the content revealed when this tab is active. It stays mounted
	// whether or not it is on screen.
	Panel Component
}

// DeckTabs renders a strip of tabs over a deck of panels.
//
// The reveal is CSS, not JS: the deck is a style.SlideDeck and shows whichever
// panel carries the widget.Current state. Every panel stays mounted — the
// state decides which one is on screen, nothing is unmounted. That is
// SlideDeck's contract, and it is why a panel may hold a form mid-edit and
// survive a trip through another tab.
//
// The same mechanism runs the application shell one scale up (layout/platformd
// drives its module panels exactly this way); this is it at widget scale.
type DeckTabs struct {
	Element // value embed — never a pointer
	Items   []Item
	// Active carries the ID of the tab on screen. Optional: nil gets a signal
	// created in Init, defaulting to Items[0].ID.
	//
	// It is injected rather than owned so the active tab can be read, set and
	// persisted from outside — a URL fragment, a saved preference, a sibling
	// control that jumps to a section.
	Active *SignalString
	// Label, when set, becomes the tab strip's aria-label. A screen reader
	// announces it as the name of the tab set.
	Label string
	// OnChange fires after Active changes, never when the active tab is
	// re-clicked.
	OnChange func(id string)
}

func (t *DeckTabs) WidgetName() widget.Name { return NameDeckTabs }
func (t *DeckTabs) WidgetKind() widget.Kind { return widget.Tabs }

func (t *DeckTabs) Init(_ Ctx) {
	if t.Active == nil {
		t.Active = NewString("")
	}
	if t.Active.Get() == "" && len(t.Items) > 0 {
		t.Active.Set(t.Items[0].ID)
	}
}

// activate moves the selection and notifies. Re-clicking the active tab is a
// no-op: it must not re-fire OnChange, which a consumer may use to reload.
func (t *DeckTabs) activate(id string) {
	if t.Active == nil || t.Active.Get() == id {
		return
	}
	t.Active.Set(id)
	if t.OnChange != nil {
		t.OnChange(id)
	}
}

// isActive is the predicate both the tab and its panel bind widget.Current to,
// so the two can never disagree about which one is showing.
func (t *DeckTabs) isActive(id string) bool {
	return t.Active != nil && t.Active.Get() == id
}

func (t *DeckTabs) Render() *Element {
	if t.Active == nil {
		t.Init(nil)
	}

	root := Div().Set(clsRoot.AsAttr())
	if len(t.Items) == 0 {
		return root
	}

	strip := Nav().Set(clsList.AsAttr()).Attr("role", "tablist")
	if t.Label != "" {
		strip.Attr("aria-label", t.Label)
	}
	deck := Div().Set(clsDeck.AsAttr())

	for _, item := range t.Items {
		item := item
		id := item.ID

		// type=button explicitly: a bare <button> defaults to type=submit, so
		// a tab set placed inside a form would submit it on every tab change.
		tab := Button().Set(clsTab.AsAttr()).
			Attr("type", "button").
			Attr("role", "tab").
			Attr("data-id", id).
			BindStateFunc(widget.Current, func() bool { return t.isActive(id) }).
			OnClick(func(Event) { t.activate(id) })

		if item.Icon != "" {
			tab.Child(item.Icon.Render(string(clsIcon)))
		}
		tab.Child(Span().Set(clsLabel.AsAttr()).Text(item.Label))

		panel := Section().Set(clsPanel.AsAttr()).
			Attr("role", "tabpanel").
			Attr("data-id", id).
			BindStateFunc(widget.Current, func() bool { return t.isActive(id) })

		if item.Panel != nil {
			panel.Child(item.Panel)
		}

		// The ARIA pair is wired through GetID(), which MINTS an id on demand —
		// never through an id this component composes from Item.ID. dom resolves
		// every handler and every signal patch by id, and claimID panics when one
		// render pass writes the same id twice: two DeckTabs on one page sharing an
		// Item.ID would have crashed the render. A minted id is unique by
		// construction, so the same Item.ID may appear in as many tab sets as the
		// screen needs.
		//
		// data-id keeps Item.ID reachable for a test or a consumer's query; it is
		// not an identifier the framework resolves.
		tab.Attr("aria-controls", panel.GetID())
		panel.Attr("aria-labelledby", tab.GetID())

		strip.Child(tab)
		deck.Child(panel)
	}

	return root.Child(strip).Child(deck)
}
