// Package targetdate is targetlist's sibling for rows that need a prominent
// leading badge — an hour, a day, anything view.Item.LeadTop/Main/Bottom
// carries — instead of a plain label. Same multi-selection mechanics as
// targetlist.
package targetdate

import (
	. "webtyp.com/dom"
	"webtyp.com/fmt"
	. "webtyp.com/html"
	"webtyp.com/view"
	"webtyp.com/widget"

	"webtyp.com/components/listselect"
)

// badgeChars mirrors targetlist's own budget — see that package's comment.
const badgeChars = 16

// NameTargetDate is the widget identity.
const NameTargetDate = widget.Name("targetdate")

const (
	PartRow        = widget.Part("row")
	PartContent    = widget.Part("content")
	PartBadge      = widget.Part("badge")
	PartLabel      = widget.Part("label")
	PartList       = widget.Part("list")
	PartLead       = widget.Part("lead")
	PartLeadStack  = widget.Part("lead-stack")
	PartLeadTop    = widget.Part("lead-top")
	PartLeadMain   = widget.Part("lead-main")
	PartLeadBottom = widget.Part("lead-bottom")
)

var (
	clsListWrap   = NameTargetDate.Root()
	clsList       = NameTargetDate.Class(PartList)
	clsRow        = NameTargetDate.Class(PartRow)
	clsContent    = NameTargetDate.Class(PartContent)
	clsLabel      = NameTargetDate.Class(PartLabel)
	clsBadge      = NameTargetDate.Class(PartBadge)
	clsLead       = NameTargetDate.Class(PartLead)
	clsLeadStack  = NameTargetDate.Class(PartLeadStack)
	clsLeadTop    = NameTargetDate.Class(PartLeadTop)
	clsLeadMain   = NameTargetDate.Class(PartLeadMain)
	clsLeadBottom = NameTargetDate.Class(PartLeadBottom)
)

// Item is view.Item — see targetlist.Item's comment for why this is an
// alias, not a copy. TargetDate is the one row that actually reads
// LeadTop/Main/Bottom.
type Item = view.Item

// rowState holds the per-record reactive text of one row. The reconciler
// reuses row nodes by key across SetItems, so static Text() at build time
// froze the first render's words into surviving nodes. Binding the words to
// signals the widget rewrites on every SetItems keeps them current while the
// node — and its listeners — survives. The id never changes for a node, so
// selection derives and the click path stay keyed on it, not on the words.
type rowState struct {
	id                    string
	label, desc           *SignalString
	top, main, bottom     *SignalString
}

func newRowState(it Item) *rowState {
	return &rowState{
		id:     it.ID,
		label:  NewString(it.Label),
		desc:   NewString(it.Description),
		top:    NewString(it.LeadTop),
		main:   NewString(it.LeadMain),
		bottom: NewString(it.LeadBottom),
	}
}

// set refreshes the words of a retained row to the latest record.
func (r *rowState) set(it Item) {
	r.label.Set(it.Label)
	r.desc.Set(it.Description)
	r.top.Set(it.LeadTop)
	r.main.Set(it.LeadMain)
	r.bottom.Set(it.LeadBottom)
}

// TargetDate is a selectable list of records with a prominent leading badge
// (LeadTop/Main/Bottom) instead of a plain label lead-in.
type TargetDate struct {
	Element

	Selected *SignalString

	OnSelect func(it Item)

	items []Item
	rows  *SignalNodes
	sel   listselect.Mode

	rowStates []*rowState // one per live record id, pruned on every SetItems
}

func (t *TargetDate) WidgetName() widget.Name { return NameTargetDate }
func (t *TargetDate) WidgetKind() widget.Kind { return widget.Combobox }

func (t *TargetDate) ensure() {
	if t.rows == nil {
		t.rows = NewNodes()
	}
	if t.Selected == nil {
		t.Selected = NewString("")
	}
}

func (t *TargetDate) Init(_ Ctx) { t.ensure() }

func (t *TargetDate) SetSelectMode(on bool)        { t.sel.SetOn(on) }
func (t *TargetDate) SetDanger(on bool)            { t.sel.SetDanger(on) }
func (t *TargetDate) OnCheckedChange(fn func(int)) { t.sel.OnChange = fn }

// itemIDs is the "current rows" listselect.Header/RowOf read to size the "k /
// N" count and the select-all tri-state. t.items is a plain field, not a
// signal — reading t.rows.Get() first is what makes a derive that calls
// itemIDs() re-run on every SetItems (a reload, a filter, a day switch), not
// just on a selection change. Skipping this read is the exact bug that left
// the header's count frozen after a reload: the derive had nothing here to
// resubscribe to.
func (t *TargetDate) itemIDs() []string {
	_ = t.rows.Get()
	ids := make([]string, len(t.items))
	for i, it := range t.items {
		ids[i] = it.ID
	}
	return ids
}

func (t *TargetDate) CheckedIDs() []string {
	return t.sel.CheckedIDs(t.itemIDs())
}

func (t *TargetDate) SetItems(items []Item) {
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
func (t *TargetDate) stateFor(it Item) *rowState {
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
func (t *TargetDate) itemByID(id string) Item {
	for _, it := range t.items {
		if it.ID == id {
			return it
		}
	}
	return Item{ID: id}
}

func (t *TargetDate) Items() []Item { return t.items }
func (t *TargetDate) Count() int    { return len(t.items) }

func (t *TargetDate) Render() *Element {
	list := Ul().Set(clsList.AsAttr()).Attr("role", "listbox").BindChildren(t.rows)
	return Div().Set(clsListWrap.AsAttr()).
		BindStateFunc(widget.Open, func() bool { return t.sel.On().Get() }).
		Child(listselect.Header(&t.sel, t.itemIDs, t.WidgetName())).
		Child(list)
}

func (t *TargetDate) buildRow(it Item) *Element {
	return t.buildRowEl(t.stateFor(it))
}

func (t *TargetDate) buildRowEl(st *rowState) *Element {
	id := st.id
	key := "td-" + id

	// RowOf owns the per-field selection wiring: the narrow Edit/Danger derives
	// and the check box. isSel widens Edit with the normal-mode highlight for
	// the ROW only; the box binds RowOf's narrow ones, so a row merely loaded
	// in normal mode reveals no glyph. See targetlist.buildRow.
	r := listselect.RowOf(&t.sel, id, t.WidgetName())
	isSel := DeriveBool(func() bool {
		_ = t.sel.Changed().Get()
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

	// Words ride signals, not static text: this shell may be discarded by the
	// reconciler in favour of a surviving node bound to the same signals, or
	// survive itself into the next render — either way the strings stay
	// current because only the signals carry them.
	lead := Div().Set(clsLead.AsAttr()).Child(
		Div().Set(clsLeadStack.AsAttr()).Child(
			Span().Set(clsLeadTop.AsAttr()).BindText(st.top),
			Span().Set(clsLeadMain.AsAttr()).BindText(st.main),
			Span().Set(clsLeadBottom.AsAttr()).BindText(st.bottom),
		),
	)

	// The box owns which glyph shows and its colour via its own
	// Selected (edit) / Invalid (delete) state, written only in selection
	// mode by RowOf. Nothing here hangs off the ROW's state.
	content := Div().Set(clsContent.AsAttr()).
		Child(r.Check).
		Child(Span().Set(clsLabel.AsAttr()).BindText(st.label))
	// The badge mounts once and hides on empty, instead of existing
	// conditionally: a conditional would need a new node exactly when the
	// description flips empty↔set (a reservation confirming), which is the
	// same staleness this change removes. Show() keeps it mounted and merely
	// unhides it.
	content.Child(Show(DeriveBool(func() bool { return st.desc.Get() != "" }),
		Span().Set(clsBadge.AsAttr()).
			BindAttr("title", st.desc).
			BindText(DeriveString(func() string {
				return fmt.Convert(st.desc.Get()).Truncate(badgeChars).String()
			}))))

	row.Child(lead)
	row.Child(content)

	return row
}
