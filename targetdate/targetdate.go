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

// TargetDate is a selectable list of records with a prominent leading badge
// (LeadTop/Main/Bottom) instead of a plain label lead-in.
type TargetDate struct {
	Element

	Selected *SignalString

	OnSelect func(it Item)

	items []Item
	rows  *SignalNodes
	sel   listselect.Mode
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
	nodes := make([]*Element, 0, len(items))
	for _, it := range items {
		nodes = append(nodes, t.buildRow(it))
	}
	t.rows.Set(nodes)
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
	id := it.ID
	key := "td-" + id

	// RowOf owns the per-row selection wiring: the narrow Edit/Danger derives
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
		Attr("role", "option").
		BindState(widget.Selected, isSel).
		BindState(widget.Invalid, r.Danger).
		BindAttrBool("aria-selected", DeriveBool(func() bool { return isSel.Get() || r.Danger.Get() }))

	row.On("click", func(Event) {
		if t.sel.On().Get() {
			t.sel.Toggle(id)
			return
		}
		if t.OnSelect != nil {
			t.OnSelect(it)
		}
	})

	lead := Div().Set(clsLead.AsAttr()).Child(
		Div().Set(clsLeadStack.AsAttr()).Child(
			Span().Set(clsLeadTop.AsAttr()).Text(it.LeadTop),
			Span().Set(clsLeadMain.AsAttr()).Text(it.LeadMain),
			Span().Set(clsLeadBottom.AsAttr()).Text(it.LeadBottom),
		),
	)

	// The box owns which glyph shows and its colour via its own
	// Selected (edit) / Invalid (delete) state, written only in selection
	// mode by RowOf. Nothing here hangs off the ROW's state.
	content := Div().Set(clsContent.AsAttr()).
		Child(r.Check).
		Child(Span().Set(clsLabel.AsAttr()).Text(it.Label))
	if it.Description != "" {
		content.Child(Span().Set(clsBadge.AsAttr()).
			Attr("title", it.Description).
			Text(fmt.Convert(it.Description).Truncate(badgeChars).String()))
	}

	row.Child(lead)
	row.Child(content)

	return row
}
