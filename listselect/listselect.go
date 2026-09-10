// Package listselect owns one concern: the multi-selection mode a record list
// enters when its host is about to act on several rows at once. It is a lego
// piece — targetlist and targetdate assemble it, they do not re-declare it.
//
// Sibling of listgap, and for the same reason: two lists that must stay
// visually and behaviourally interchangeable (crudview swaps one for the
// other) cannot each own a private copy of the rule.
package listselect

import (
	. "webtyp.com/dom"
	"webtyp.com/fmt"
	"webtyp.com/html"
	"webtyp.com/icons/pencil"
	"webtyp.com/icons/selectall"
	"webtyp.com/icons/trash"
	"webtyp.com/widget"
)

// Mode is the selection state of one list. The zero value is a usable list in
// normal mode — a list is a list until its host says otherwise.
type Mode struct {
	on *SignalBool

	// checked is a SLICE, never a map. A map pulls TinyGo's hashing and
	// runtime machinery into the WASM binary, and this type compiles to the
	// browser. A record list holds tens of rows, so a linear scan costs
	// microseconds — the map was never buying anything here.
	//
	// []string and not []fmt.KeyValue: this is a SET, so there is no value to
	// carry. fmt.KeyValue is the map-free shape for an actual key→value pair;
	// inventing a Value of "true" would be a stringly-typed bool.
	checked []string

	// changed flips on every Toggle. checked is a plain slice — reading it
	// inside a DeriveBool tracks nothing, so without this signal a row's
	// bound states would go stale the moment after selection mode opens and
	// no tap would ever repaint. Rows read Changed() first in their
	// derives; the flip re-runs them and IsChecked is re-read live.
	changed *SignalBool

	// danger arms the danger tone: while set, a checked row means "marked
	// for a destructive action" and the skin paints it red instead of blue.
	// Owned by the host (crudview sets it from its mode); a plain list that
	// never arms it keeps the single Accent selection language it always had.
	danger *SignalBool

	// OnChange fires after every toggle with the current count, so a host can
	// label its commit button ("🗑 3") and disable it at zero.
	OnChange func(n int)
}

func (m *Mode) ensure() {
	if m.on == nil {
		m.on = NewBool(false)
	}
}

// On reports the signal a component binds its root state to, so the stylesheet
// can reveal the checks. Never a bool: the skin has to react.
func (m *Mode) On() *SignalBool {
	m.ensure()
	return m.on
}

// SetDanger arms or disarms the danger tone. Additive: a host that never
// calls it gets no red anywhere, whatever the selection does.
func (m *Mode) SetDanger(on bool) {
	m.ensure()
	if m.danger == nil {
		m.danger = NewBool(false)
	}
	m.danger.Set(on)
}

// Danger reports the tone signal a row binds its Invalid state to. Never a
// bool: the skin has to react.
func (m *Mode) Danger() *SignalBool {
	m.ensure()
	if m.danger == nil {
		m.danger = NewBool(false)
	}
	return m.danger
}

// Changed flips on every Toggle. Rows read it first in their state derives
// (see the changed field); without that read a tap would update nothing on
// screen.
func (m *Mode) Changed() *SignalBool {
	m.ensure()
	if m.changed == nil {
		m.changed = NewBool(false)
	}
	return m.changed
}

// SetOn enters or leaves selection mode. Leaving ALWAYS clears the marks:
// a mode the user cancelled must not leave a hidden selection behind for the
// next entry to inherit silently.
func (m *Mode) SetOn(on bool) {
	m.ensure()
	if !on && len(m.checked) > 0 {
		m.checked = nil
		if m.OnChange != nil {
			m.OnChange(0)
		}
	}
	m.on.Set(on)
}

// Toggle marks or unmarks one id and fires OnChange.
func (m *Mode) Toggle(id string) {
	m.ensure()
	found := -1
	for i, c := range m.checked {
		if c == id {
			found = i
			break
		}
	}

	if found >= 0 {
		// remove
		m.checked = append(m.checked[:found], m.checked[found+1:]...)
	} else {
		// add
		m.checked = append(m.checked, id)
	}
	m.Changed().Toggle()

	if m.OnChange != nil {
		m.OnChange(len(m.checked))
	}
}

// CheckAll marks every id in ids — the caller's CURRENT render order — and
// replaces any previous selection. It owns a fresh backing array (never
// aliases the caller's slice). Fires Changed() and OnChange with the new
// count. This is the master check's "select all" action.
func (m *Mode) CheckAll(ids []string) {
	m.ensure()
	m.checked = append([]string(nil), ids...)
	m.Changed().Toggle()
	if m.OnChange != nil {
		m.OnChange(len(m.checked))
	}
}

// Clear unmarks every row WITHOUT leaving selection mode — unlike
// SetOn(false), which also exits the mode. The master check's "deselect all".
// A no-op (no signal churn) when nothing is marked.
func (m *Mode) Clear() {
	m.ensure()
	if len(m.checked) == 0 {
		return
	}
	m.checked = nil
	m.Changed().Toggle()
	if m.OnChange != nil {
		m.OnChange(0)
	}
}

// Count reports how many rows are currently marked. The master check reads it
// to decide its tri-state (none / some / all) and to render "n / total".
func (m *Mode) Count() int { return len(m.checked) }

// IsChecked answers for one id — what a row binds its check state to.
func (m *Mode) IsChecked(id string) bool {
	for _, c := range m.checked {
		if c == id {
			return true
		}
	}
	return false
}

// CheckedIDs returns the marked ids in the order given by ids, which the
// caller passes as its current render order.
//
// Ordering is NOT optional and NOT the caller's problem to remember: checked
// accumulates in TAP order, and a host building a confirmation message from
// tap order would list rows in an order that matches nothing on screen.
// Taking the render order as a parameter is what makes the wrong version
// unwritable — there is no accessor that returns the raw slice.
func (m *Mode) CheckedIDs(ids []string) []string {
	if len(m.checked) == 0 {
		return nil
	}
	var res []string
	for _, id := range ids {
		if m.IsChecked(id) {
			res = append(res, id)
		}
	}
	return res
}

// Part names for the selection chrome. Unexported: the host passes only its
// widget Name, and listselect derives the classes from it, so neither the
// CSS (ApplyRow/ApplyHeader) nor the markup (RowOf/Header) can drift from a
// name the host types by hand.
const (
	partHeader      = widget.Part("sel-header")
	partCheck       = widget.Part("sel-check")
	partCheckTrash  = widget.Part("sel-check-trash")
	partCheckPencil = widget.Part("sel-check-pencil")
	partAll         = widget.Part("sel-all")
	partAllIcon     = widget.Part("sel-all-icon")
	partAllCount    = widget.Part("sel-count")
	partAllSpacer   = widget.Part("sel-spacer")
)

// Row is the per-row selection wiring listselect hands a target* widget so the
// three widgets stop hand-rolling identical derives. Build it once per row in
// buildRow.
//
//   - Check is the glyph box: place it in the row. It reveals a trash glyph
//     when the row is marked while the danger tone is armed, a pencil when
//     marked while it is not, and is invisible otherwise.
//   - Edit   is "marked, danger tone OFF" — the widget ORs this with its own
//     "this is the loaded record" highlight for the row's Selected state.
//   - Danger is "marked, danger tone ON" — bind the row's Invalid state to it.
type Row struct {
	Check  *Element
	Edit   *SignalBool
	Danger *SignalBool
}

// RowOf builds the selection wiring for one row id. name is the host widget's
// WidgetName() — listselect namespaces its parts under it so the CSS
// (ApplyRow) and the element agree.
func RowOf(m *Mode, id string, name widget.Name) Row {
	edit := DeriveBool(func() bool {
		_ = m.Changed().Get() // re-read IsChecked after every tap (see Mode.Changed)
		return m.On().Get() && !m.Danger().Get() && m.IsChecked(id)
	})
	danger := DeriveBool(func() bool {
		_ = m.Changed().Get()
		return m.On().Get() && m.Danger().Get() && m.IsChecked(id)
	})
	check := html.Span().Set(name.Class(partCheck).AsAttr()).
		BindState(widget.Selected, edit).
		BindState(widget.Invalid, danger).
		Child(trash.Ref.Render(string(name.Class(partCheckTrash)))).
		Child(pencil.Ref.Render(string(name.Class(partCheckPencil))))
	return Row{Check: check, Edit: edit, Danger: danger}
}

// Header builds the in-flow selection header strip: a select-all /
// deselect-all box and a count that reads "k / N". Child() it above the
// list <ul>.
//
// The box always carries the same selectall glyph — never trash or pencil.
// Those name the ACTION the marked rows are about to feed (delete, bulk
// edit); the box's own job is the SELECTION, not the action, and a host
// already shows the action glyph on its own commit button (crudview's
// footer 🗑/✏). Painting that same glyph here duplicated it. The box's
// background still tracks the danger tone (Danger red / Accent amber, via
// its Invalid/Selected state below) — only the glyph on top stays fixed.
//
// Both states additionally require Count() > 0: painting the box the moment
// selection mode opens, before anything is checked, gave tapping
// select-all no visible effect — the box already looked "active". Gating on
// an actual mark makes the tap read as a real state change: resting/Inset
// at zero, Danger/Accent from one mark onward.
//
// The strip itself always reserves its row height — never Hide()'s — so the
// list never shifts when selection mode opens or closes. The count is
// ALWAYS visible (see ApplyHeader); only the box is hidden in normal mode
// and revealed on the root's Open state. The spacer stays visible always,
// sized like the box, which is what reserves the strip's height AND keeps
// the count centred in both modes — the count lands in the strip's true
// visual center instead of the center of whatever space is left once the
// box appears; the box itself rides the trailing edge (ApplyHeader's
// PushEnd), away from the count.
//
// ids returns the current rows in render order; the count is its length.
// name is the host's WidgetName().
func Header(m *Mode, ids func() []string, name widget.Name) *Element {
	allChecked := DeriveBool(func() bool {
		_ = m.Changed().Get()
		n := m.Count()
		return n > 0 && n == len(ids())
	})

	box := html.Span().Set(name.Class(partAll).AsAttr()).
		Attr("role", "checkbox").
		BindAttrBool("aria-checked", allChecked).
		BindState(widget.Invalid, DeriveBool(func() bool {
			_ = m.Changed().Get()
			return m.On().Get() && m.Danger().Get() && m.Count() > 0
		})).
		BindState(widget.Selected, DeriveBool(func() bool {
			_ = m.Changed().Get()
			return m.On().Get() && !m.Danger().Get() && m.Count() > 0
		})).
		Child(selectall.Ref.Render(string(name.Class(partAllIcon))))

	box.OnClick(func(Event) {
		if n := m.Count(); n > 0 && n == len(ids()) {
			m.Clear()
			return
		}
		m.CheckAll(ids())
	})

	count := html.Span().Set(name.Class(partAllCount).AsAttr()).
		BindTextFunc(func() string {
			_ = m.Changed().Get()
			return fmt.Sprintf("%d / %d", m.Count(), len(ids()))
		})

	// Balances the box's footprint on the leading edge — no click handler,
	// no text — so the count centers in the whole strip, not in whatever is
	// left once the box claims the trailing edge.
	spacer := html.Span().Set(name.Class(partAllSpacer).AsAttr())

	return html.Span().Set(name.Class(partHeader).AsAttr()).
		Child(spacer).
		Child(count).
		Child(box)
}
