package searchbar

import (
	. "webtyp.com/dom"
	. "webtyp.com/html"
	"webtyp.com/svg"
	"webtyp.com/widget"
)

// NameSearchBar is the widget identity; it produces the class prefix
// "searchbar" and the part classes "searchbar__icon", "searchbar__glyph",
// "searchbar__input".
const NameSearchBar = widget.Name("searchbar")

const (
	PartIcon  = widget.Part("icon")  // the square coloured cap holding the magnifier
	PartInput = widget.Part("input") // the text field, the body of the bar
)

var (
	clsSearchBar = NameSearchBar.Root()
	clsIcon      = NameSearchBar.Class(PartIcon)
	clsInput     = NameSearchBar.Class(PartInput)
)

// iconMagnifier is registered by IconSvg in svg.go.
const iconMagnifier = svg.Icon("searchbar-magnifier")

// defaultPlaceholder is what the field says when the host sets none.
const defaultPlaceholder = "Search…"

// SearchBar is a single-control filter bar: a magnifier cap followed by a text
// field. It holds no list and knows nothing about what it filters — it reports
// the term and the host decides what that means.
//
// It satisfies widget.Filterable, which is the whole reason it is swappable: a
// host holds the seam, not this type, so a calendar or a select that also
// implements Filterable takes its place with no change to the host.
type SearchBar struct {
	Element // value embed — NEVER *dom.Element (TinyGo heap constraint)

	// Placeholder is the field's placeholder text. Empty uses defaultPlaceholder.
	Placeholder string

	onFilter func(term string)
}

// Compile-time proof that the seam is satisfied. Keep this line: if the
// upstream interface ever changes shape, this is what fails, here, instead of
// failing at a host's type assertion where it would silently evaluate false and
// leave the bar wired to nothing.
var _ widget.Filterable = (*SearchBar)(nil)

func (s *SearchBar) WidgetName() widget.Name { return NameSearchBar }

// Form, not Combobox: the bar has no popup and no options list. Combobox would
// license Open/Selected states this control can never be in.
func (s *SearchBar) WidgetKind() widget.Kind { return widget.Form }

// OnFilterChange implements widget.Filterable: it registers the sink for every
// keystroke. The host calls it once while wiring its filter slot; passing nil
// clears the sink.
//
// The signature is fixed by widget.Filterable — do not add a parameter, do not
// return anything, do not add a companion getter.
func (s *SearchBar) OnFilterChange(fn func(term string)) { s.onFilter = fn }

func (s *SearchBar) Render() *Element {
	root := Div().Set(clsSearchBar.AsAttr())

	// A <label> so a click on the cap focuses nothing accidentally and the
	// magnifier is announced as decoration, not as a button.
	root.Child(Label().Set(clsIcon.AsAttr()).
		Child(iconMagnifier.Render()))

	placeholder := s.Placeholder
	if placeholder == "" {
		placeholder = defaultPlaceholder
	}

	// size="1" collapses the input's intrinsic width (the UA default is 20
	// characters): the Row flexbox wraps on the flex base size, and ~240px of
	// intrinsic width made the bar wrap onto two lines inside narrow hosts. The
	// flex Grow still sizes the rendered field; size only feeds the intrinsic
	// measurement.
	//
	// type="search" carries its own clear control — the browser paints and
	// wires the ✕. The bar does not add one: a second ✕ is the thing this
	// note exists to prevent. Clearing it fires an "input" event, so the
	// filter re-runs with "" through the same handler.
	input := Input("search").Set(clsInput.AsAttr()).
		Attr("placeholder", placeholder).
		Attr("size", "1")
	input.OnInput(func(e Event) {
		if s.onFilter != nil {
			s.onFilter(e.TargetValue())
		}
	})
	root.Child(input)

	return root
}
