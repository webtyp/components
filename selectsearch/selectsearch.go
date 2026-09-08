package selectsearch

import (
	. "webtyp.com/dom"
	"webtyp.com/fmt"
	. "webtyp.com/html"
	"webtyp.com/svg"
	"webtyp.com/widget"
)

// NameSelectSearch is the widget name.
const NameSelectSearch = widget.Name("selectsearch")

const (
	PartToggle      = widget.Part("toggle")
	PartBackdrop    = widget.Part("backdrop")
	PartDropdown    = widget.Part("dropdown")
	PartHeader      = widget.Part("header")
	PartHeaderBody  = widget.Part("header-body")
	PartPlaceholder = widget.Part("placeholder")
	PartIcon        = widget.Part("icon")
	PartGlyph       = widget.Part("glyph")
	PartSearch      = widget.Part("search")
	PartOptions     = widget.Part("options")
	PartOption      = widget.Part("option")
	PartText        = widget.Part("text")
	PartLabel       = widget.Part("label")
	PartSublabel    = widget.Part("sublabel")
	PartDesc        = widget.Part("desc")
)

var (
	ClsSsBox         = NameSelectSearch.Root()
	ClsSsToggle      = NameSelectSearch.Class(PartToggle)
	ClsSsBackdrop    = NameSelectSearch.Class(PartBackdrop)
	ClsSsDropdown    = NameSelectSearch.Class(PartDropdown)
	ClsSsHeader      = NameSelectSearch.Class(PartHeader)
	ClsSsHeaderBody  = NameSelectSearch.Class(PartHeaderBody)
	ClsSsPlaceholder = NameSelectSearch.Class(PartPlaceholder)
	ClsSsIcon        = NameSelectSearch.Class(PartIcon)
	ClsSsGlyph       = NameSelectSearch.Class(PartGlyph)
	ClsSsSearch      = NameSelectSearch.Class(PartSearch)
	ClsSsOptions     = NameSelectSearch.Class(PartOptions)
	ClsSsOption      = NameSelectSearch.Class(PartOption)
	ClsSsText        = NameSelectSearch.Class(PartText)
	ClsSsLabel       = NameSelectSearch.Class(PartLabel)
	ClsSsSublabel    = NameSelectSearch.Class(PartSublabel)
	ClsSsDesc        = NameSelectSearch.Class(PartDesc)
)

const iconArrowDown = svg.Icon("ss-arrow-down")

// SearchMode decides whether the picker shows its search field.
//
// The field is not free: on a phone it summons the on-screen keyboard the
// instant the control opens, covering half the list the user came to read.
// That is a good trade when the list is long enough that scanning it is
// slower than typing, and a bad one when it holds five names.
type SearchMode uint8

const (
	// SearchAuto shows the field only when the list is long enough to be
	// worth filtering, or when an OnSearch source is wired — a consumer that
	// fetches results for a term has already declared the list is not
	// browsable, whatever len(Options) says at this instant.
	//
	// It is the zero value on purpose: it is the answer that is right without
	// the consumer having to think, and the one that is right for the case
	// that actually broke (a handful of patients, keyboard in the way).
	SearchAuto SearchMode = iota
	// SearchAlways keeps the field however short the list is.
	SearchAlways
	// SearchNever drops it however long the list is.
	SearchNever
)

// searchThreshold is the option count from which SearchAuto decides typing
// beats scanning. Calibrated against what the sheet can actually show: at
// Capped(ExtentMost) a phone fits roughly six rows, so a list that needs more
// than one screenful is one the user cannot take in at a glance.
const searchThreshold = 10

// SsOption represents a selectable item.
type SsOption struct {
	ID          string // unique identifier, returned in OnSelect
	Label       string // visible text
	Sublabel    string // optional second line under Label — position only, no assumed content
	Description string // optional badge shown on the right
}

type SelectSearch struct {
	Element                                  // value embed — NEVER pointer (TinyGo heap constraint)
	Placeholder string                       // text shown when nothing is selected
	Options     []SsOption                   // initial static options
	Search      SearchMode                   // whether the search field appears; zero value is SearchAuto
	OnSelect    func(id, description string) // called when user picks an option
	OnSearch    func(term string) []SsOption // called when ALL local options are filtered out

	// Internal state signals
	selectedLabel *SignalString
	selectedID    *SignalString
	// The rest of the picked option, held so the collapsed header can show the
	// same three-part layout an open row does — name over id, time on the
	// right — instead of a single flattened line. Written only in selectOption,
	// alongside selectedLabel/selectedID.
	selectedSublabel *SignalString
	selectedDesc     *SignalString
	query            *SignalString
	isOpen           *SignalBool
	rows             *SignalNodes
	searchShown      *SignalBool

	onFilter func(term string) // set via OnFilterChange — satisfies widget.Filterable
}

func (c *SelectSearch) WidgetName() widget.Name { return NameSelectSearch }
func (c *SelectSearch) WidgetKind() widget.Kind { return widget.Combobox }

var _ widget.Filterable = (*SelectSearch)(nil)

// OnFilterChange implements widget.Filterable: it registers the sink called
// with the picked option's ID whenever a selection is made. This is a
// SEPARATE, additive wiring path from OnSelect — OnSelect still gets
// (id, description) for a consumer that needs both; OnFilterChange exists so
// a host that only knows the generic Filterable contract (e.g.
// webtyp/layout/crudview's Filter slot) can drop a *SelectSearch into the
// same seam a *searchbar.SearchBar fills today, with no bespoke glue.
//
// The signature is fixed by widget.Filterable — do not add a parameter, do
// not return anything, do not add a companion getter (see searchbar.go's
// OnFilterChange for the same rule stated for SearchBar).
func (c *SelectSearch) OnFilterChange(fn func(term string)) { c.onFilter = fn }

func (c *SelectSearch) Init(_ Ctx) {
	c.selectedLabel = NewString("")
	c.selectedID = NewString("")
	c.selectedSublabel = NewString("")
	c.selectedDesc = NewString("")
	c.query = NewString("")
	c.isOpen = NewBool(false)
	c.rows = NewNodes(c.buildRows("")...)
	c.searchShown = NewBool(c.searchVisible())
}

// searchVisible resolves SearchMode against the list as it stands right now.
// It is a plain function, not a derived signal: Options is an ordinary slice,
// so nothing would re-run it on a change. SetOptions pushes the result into
// searchShown instead — the two write sites are Init and SetOptions, and
// there is no third way for the option list to move.
func (c *SelectSearch) searchVisible() bool {
	switch c.Search {
	case SearchAlways:
		return true
	case SearchNever:
		return false
	default:
		return c.OnSearch != nil || len(c.Options) >= searchThreshold
	}
}

// SetOptions replaces the option list — safe to call after Init/Render,
// e.g. once options from an async source (fetch, MCP call) arrive.
// Preserves the current search query filter, if any.
func (c *SelectSearch) SetOptions(options []SsOption) {
	c.Options = options
	c.rows.Set(c.buildRows(c.query.Get()))
	// The count is what SearchAuto decides on, so a list that arrives from an
	// async source has to be able to move that decision. Without this line a
	// picker rendered empty and filled later would never grow its field.
	c.searchShown.Set(c.searchVisible())
}

func (c *SelectSearch) Render() *Element {
	placeholderText := c.Placeholder
	if placeholderText == "" {
		placeholderText = "Select..."
	}

	hasSelection := DeriveBool(func() bool { return c.selectedLabel.Get() != "" })
	noSelection := DeriveBool(func() bool { return c.selectedLabel.Get() == "" })
	hasSublabel := DeriveBool(func() bool { return c.selectedSublabel.Get() != "" })
	hasDesc := DeriveBool(func() bool { return c.selectedDesc.Get() != "" })

	searchInput := Input("search").
		Set(ClsSsSearch.AsAttr()).
		Key("search").
		Attr("placeholder", "Search...").
		Attr("role", "combobox").
		BindAttrBool("aria-expanded", c.isOpen).
		Bind(c.query)

	toggle := Input("checkbox").Set(ClsSsToggle.AsAttr()).
		Key("toggle").
		BindAttrBool("checked", c.isOpen).
		On("change", func(e Event) {
			checked := e.TargetChecked()
			c.isOpen.Set(checked)
			if checked && c.searchShown.Get() {
				if ref, ok := searchInput.Ref(); ok {
					ref.Focus()
				}
			}
		})

	icon := Div().Set(ClsSsIcon.AsAttr()).Child(iconArrowDown.Render(string(ClsSsGlyph)))

	pickedText := Div().Set(ClsSsText.AsAttr()).
		Child(Span().Set(ClsSsLabel.AsAttr()).BindText(c.selectedLabel)).
		Child(Show(hasSublabel, Span().Set(ClsSsSublabel.AsAttr()).BindText(c.selectedSublabel)))

	headerBody := Div().Set(ClsSsHeaderBody.AsAttr()).
		Child(Show(noSelection, Span().Set(ClsSsPlaceholder.AsAttr()).Text(placeholderText))).
		Child(Show(hasSelection, pickedText)).
		Child(Show(hasDesc, Span().Set(ClsSsDesc.AsAttr()).BindText(c.selectedDesc)))

	header := Label().Set(ClsSsHeader.AsAttr()).
		For(toggle).
		Child(icon).
		Child(headerBody)

	optList := Ul().Set(ClsSsOptions.AsAttr()).
		Key("options").
		Attr("role", "listbox").
		BindChildren(c.rows)

	searchInput.
		Attr("aria-controls", optList.GetID()).
		On("input", func(e Event) {
			term := e.TargetValue()

			if term != "" {
				allHidden := true
				for _, opt := range c.Options {
					if fmt.Matches(opt.Label, term) || fmt.Matches(opt.Description, term) {
						allHidden = false
						break
					}
				}
				if allHidden && c.OnSearch != nil {
					if newOpts := c.OnSearch(term); len(newOpts) > 0 {
						c.Options = append(c.Options, newOpts...)
					}
				}
			}

			c.rows.Set(c.buildRows(term))
		})

	dropdown := Div().Set(ClsSsDropdown.AsAttr()).
		Child(Show(c.searchShown, searchInput)).
		Child(optList)

	backdrop := Div().Set(ClsSsBackdrop.AsAttr()).
		On("click", func(e Event) { c.isOpen.Set(false) })

	return Div().Set(ClsSsBox.AsAttr()).
		BindState(widget.Open, c.isOpen).
		Child(toggle).
		Child(header).
		Child(Show(c.isOpen, backdrop)).
		Child(Show(c.isOpen, dropdown))
}

// selectOption is the single place an option becomes "chosen" — today only
// a mouse click reaches it, but every future input path (keyboard, a future
// OnSearch auto-pick) commits through here too, so OnSelect and the
// Filterable sink can never fire out of step with each other.
func (c *SelectSearch) selectOption(o SsOption) {
	c.selectedLabel.Set(o.Label)
	c.selectedID.Set(o.ID)
	c.selectedSublabel.Set(o.Sublabel)
	c.selectedDesc.Set(o.Description)
	c.isOpen.Set(false)
	c.query.Set("")
	c.rows.Set(c.buildRows(""))
	if c.OnSelect != nil {
		c.OnSelect(o.ID, o.Description)
	}
	if c.onFilter != nil {
		c.onFilter(o.ID)
	}
}

func (c *SelectSearch) buildRows(term string) []*Element {
	var rows []*Element
	for _, opt := range c.Options {
		if term != "" && !fmt.Matches(opt.Label, term) && !fmt.Matches(opt.Description, term) {
			continue
		}

		o := opt // capture loop variable

		// text stacks Label over the optional Sublabel — a second line under
		// the name, not a sibling beside it, which is why it is its own Grow()
		// column instead of two more spans loose in the row's own Row().
		text := Div().Set(ClsSsText.AsAttr()).
			Child(Span().Set(ClsSsLabel.AsAttr()).Text(opt.Label))
		if opt.Sublabel != "" {
			text.Child(Span().Set(ClsSsSublabel.AsAttr()).Text(opt.Sublabel))
		}

		item := Li().Set(ClsSsOption.AsAttr()).
			Key(opt.ID).
			Attr("role", "option").
			BindStateFunc(widget.Selected, func() bool { return c.selectedID.Get() == o.ID }).
			Child(text).
			On("click", func(e Event) { c.selectOption(o) })

		if opt.Description != "" {
			item.Child(Span().Set(ClsSsDesc.AsAttr()).Text(opt.Description))
		}
		rows = append(rows, item)
	}
	return rows
}
