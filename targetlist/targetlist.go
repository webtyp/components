// Package targetlist is the selectable record list used by CRUD views.
package targetlist

import (
	. "webtyp.com/dom"
	"webtyp.com/fmt"
	. "webtyp.com/html"
	"webtyp.com/view"
	"webtyp.com/widget"

	"webtyp.com/components/listselect"
)

// badgeChars is the badge's budget, calibrated against the --chip-width the
// skin gives it. Truncate counts the three-byte ellipsis inside this number.
const badgeChars = 16

// NameTargetList is the widget identity.
const NameTargetList = widget.Name("targetlist")

const (
	PartRow   = widget.Part("row")
	PartBadge = widget.Part("badge")
	PartLabel = widget.Part("label")
	PartList  = widget.Part("list")
)

var (
	clsListWrap = NameTargetList.Root()
	clsList     = NameTargetList.Class(PartList)
	clsRow      = NameTargetList.Class(PartRow)
	clsLabel    = NameTargetList.Class(PartLabel)
	clsBadge    = NameTargetList.Class(PartBadge)
)

// Item is view.Item, not a copy: a shared shape means crudview.filter's
// []view.Item flows straight into SetItems, and a host swapping this widget
// for targetdate (also view.Item-based) needs no re-mapping either. TargetList
// itself only ever reads ID/Label/Description — LeadTop/Main/Bottom are
// targetdate's slot, ignored here.
type Item = view.Item

// rowState holds the per-record reactive text of one row. The reconciler
// reuses row nodes by key across SetItems, so static Text() at build time
// froze the first render's words into surviving nodes. Binding the words to
// signals the widget rewrites on every SetItems keeps them current while the
// node — and its listeners — survives. The id never changes for a node, so
// selection derives and the click path stay keyed on it, not on the words.
type rowState struct {
	id          string
	label, desc *SignalString
}

func newRowState(it Item) *rowState {
	return &rowState{
		id:    it.ID,
		label: NewString(it.Label),
		desc:  NewString(it.Description),
	}
}

// set refreshes the words of a retained row to the latest record.
func (r *rowState) set(it Item) {
	r.label.Set(it.Label)
	r.desc.Set(it.Description)
}

// TargetList is a selectable list of records with a multi-selection mode.
type TargetList struct {
	Element

	// Selected holds the id of the highlighted row. Optional — created if nil so a
	// host can share it (e.g. a CRUD view binding the form to the same signal).
	Selected *SignalString

	// Row callbacks. Optional.
	OnSelect func(it Item) // row body clicked

	items []Item
	rows  *SignalNodes
	sel   listselect.Mode

	rowStates []*rowState // one per live record id, pruned on every SetItems
}

func (t *TargetList) WidgetName() widget.Name { return NameTargetList }
func (t *TargetList) WidgetKind() widget.Kind { return widget.Combobox }

func (t *TargetList) ensure() {
	if t.rows == nil {
		t.rows = NewNodes()
	}
	if t.Selected == nil {
		t.Selected = NewString("")
	}
}

func (t *TargetList) Init(_ Ctx) { t.ensure() }

func (t *TargetList) SetSelectMode(on bool)        { t.sel.SetOn(on) }
func (t *TargetList) SetDanger(on bool)            { t.sel.SetDanger(on) }
func (t *TargetList) OnCheckedChange(fn func(int)) { t.sel.OnChange = fn }

// itemIDs is the "current rows" listselect.Header/RowOf read to size the "k /
// N" count and the select-all tri-state. t.items is a plain field, not a
// signal — reading t.rows.Get() first is what makes a derive that calls
// itemIDs() re-run on every SetItems (a reload, a filter, a day switch in a
// calendar-backed host), not just on a selection change. Skipping this read
// is the exact bug that left the header's count frozen after a reload: the
// derive had nothing here to resubscribe to.
func (t *TargetList) itemIDs() []string {
	_ = t.rows.Get()
	ids := make([]string, len(t.items))
	for i, it := range t.items {
		ids[i] = it.ID
	}
	return ids
}

func (t *TargetList) CheckedIDs() []string {
	return t.sel.CheckedIDs(t.itemIDs())
}

// SetItems replaces the visible rows. Safe to call from a host on every filter or
// reload; rows go through the keyed reconcile so their bindings stay wired.
func (t *TargetList) SetItems(items []Item) {
	t.ensure()
	t.items = items
	// Retained states, pruned to the live ids: a dropped record's signals go
	// with it (its node is unmounted by the reconciler), so nothing leaks
	// across reloads and no closure outlives its record.
	var kept []*rowState
	nodes := make([]*Element, 0, len(items))
	for _, it := range items {
		st := t.stateFor(it)
		kept = append(kept, st)
		nodes = append(nodes, t.buildRowEl(st))
	}
	t.rowStates = kept
	t.rows.Set(nodes)
}

// stateFor returns the retained row state for id, refreshing its words — or
// mints it for a record the list has never shown. Linear scan, no map: a
// projected list holds tens of rows.
func (t *TargetList) stateFor(it Item) *rowState {
	for _, r := range t.rowStates {
		if r.id == it.ID {
			r.set(it)
			return r
		}
	}
	return newRowState(it)
}

// itemByID resolves the CURRENT record for a row click. A reused node's
// closure must not ship the struct its node was built with — the words may
// have moved on while the node survived. The id is immutable per node, so it
// is the lookup key; the fallback keeps the total function honest for an id
// with no row (unreachable: clicks only fire for rendered rows).
func (t *TargetList) itemByID(id string) Item {
	for _, it := range t.items {
		if it.ID == id {
			return it
		}
	}
	return Item{ID: id}
}

// Items returns the current items (the data behind the rendered rows).
func (t *TargetList) Items() []Item { return t.items }

// Count reports how many rows are currently rendered (used by hosts/tests).
func (t *TargetList) Count() int { return len(t.items) }

func (t *TargetList) Render() *Element {
	list := Ul().Set(clsList.AsAttr()).Attr("role", "listbox").BindChildren(t.rows)

	return Div().Set(clsListWrap.AsAttr()).
		BindStateFunc(widget.Open, func() bool { return t.sel.On().Get() }).
		Child(listselect.Header(&t.sel, t.itemIDs, t.WidgetName())).
		Child(list)
}

func (t *TargetList) buildRow(it Item) *Element {
	return t.buildRowEl(t.stateFor(it))
}

func (t *TargetList) buildRowEl(st *rowState) *Element {
	id := st.id
	key := "tl-" + id

	// RowOf owns the per-row selection wiring: the narrow Edit/Danger
	// derives and the check box. isSel widens Edit with the normal-mode
	// "loaded record" highlight for the ROW's fill; the box binds RowOf's
	// narrow ones, so a row merely loaded in normal mode never reveals a
	// glyph. Selected and Invalid never coincide on one element: a checked
	// row under the armed danger tone is Invalid (red), otherwise Selected
	// (blue) — one element, one fill, no race in the cascade.
	r := listselect.RowOf(&t.sel, id, t.WidgetName())
	isSel := DeriveBool(func() bool {
		_ = t.sel.Changed().Get() // re-read after every tap (see Mode.Changed)
		if t.sel.On().Get() {
			return r.Edit.Get()
		}
		return t.Selected.Get() == id
	})

	row := Li().Set(clsRow.AsAttr()).
		Key(key).
		Attr("data-row", key).
		Attr("role", "option").
		BindState(widget.Selected, isSel).
		BindState(widget.Invalid, r.Danger).
		BindAttrBool("aria-selected", DeriveBool(func() bool { return isSel.Get() || r.Danger.Get() }))

	row.OnClick(func(Event) {
		if t.sel.On().Get() {
			t.sel.Toggle(id)
			return
		}
		if t.OnSelect != nil {
			t.OnSelect(t.itemByID(id))
		}
	})

	row.Child(r.Check)
	row.Child(Span().Set(clsLabel.AsAttr()).BindText(st.label))
	// The badge mounts once and hides on empty, instead of existing
	// conditionally: a conditional would need a new node exactly when the
	// description flips empty↔set, which is the same staleness this change
	// removes. Show() keeps it mounted and merely unhides it.
	row.Child(Show(DeriveBool(func() bool { return st.desc.Get() != "" }),
		Span().Set(clsBadge.AsAttr()).
			BindAttr("title", st.desc).
			BindText(DeriveString(func() string {
				return fmt.Convert(st.desc.Get()).Truncate(badgeChars).String()
			}))))

	return row
}
