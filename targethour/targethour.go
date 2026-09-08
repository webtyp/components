// Package targethour is targetlist's sibling for a day's booked slots: each
// row leads with a prominent hour (HH:MM) and may carry a status tint
// (pending / confirmed / attended). Same multi-selection mechanics as
// targetlist/targetdate — it assembles components/listselect, it does not
// re-declare it.
package targethour

import (
	. "webtyp.com/dom"
	"webtyp.com/fmt"
	. "webtyp.com/html"
	"webtyp.com/view"
	"webtyp.com/widget"

	"webtyp.com/components/listselect"
)

const badgeChars = 16

const NameTargetHour = widget.Name("targethour")

const (
	PartRow     = widget.Part("row")
	PartContent = widget.Part("content")
	PartHour    = widget.Part("hour")
	PartLabel   = widget.Part("label")
	PartBadge   = widget.Part("badge")
	PartList    = widget.Part("list")
	PartFree    = widget.Part("free")
	PartFreeAdd = widget.Part("free-add")
)

var (
	clsListWrap = NameTargetHour.Root()
	clsList     = NameTargetHour.Class(PartList)
	clsRow      = NameTargetHour.Class(PartRow)
	clsContent  = NameTargetHour.Class(PartContent)
	clsHour     = NameTargetHour.Class(PartHour)
	clsLabel    = NameTargetHour.Class(PartLabel)
	clsBadge    = NameTargetHour.Class(PartBadge)
	clsFree     = NameTargetHour.Class(PartFree)
	clsFreeAdd  = NameTargetHour.Class(PartFreeAdd)
)

type Item = view.Item

// Status is a row's booking state — drives the tint only. The zero value
// (StatusPending) paints no tint, exactly like a plain targetlist row.
type Status uint8

const (
	StatusPending   Status = iota // no tint
	StatusConfirmed               // confirmed by reception
	StatusAttended                // patient already attended
)

type TargetHour struct {
	Element

	Selected *SignalString
	OnSelect func(it Item)

	// StatusOf maps a row to its booking state for the tint. Optional — nil
	// means every row is StatusPending (no tint). The host owns the mapping
	// from its own model / view.Item to this typed enum, so the library holds
	// no localized status strings.
	StatusOf func(it Item) Status

	// FreeSlots son horas "HH:MM" reservables (sin reserva). Se renderizan como
	// filas ligeras al final de la lista, visualmente distintas de un Item real
	// (sin estado, con un "+" y marco punteado). Opcional: nil = ninguna.
	FreeSlots []string
	// OnPickFree se invoca al hacer clic en un hueco libre, con su "HH:MM".
	OnPickFree func(hhmm string)

	items []Item
	rows  *SignalNodes
	sel   listselect.Mode
}

func (t *TargetHour) WidgetName() widget.Name { return NameTargetHour }
func (t *TargetHour) WidgetKind() widget.Kind { return widget.Combobox }

func (t *TargetHour) ensure() {
	if t.rows == nil {
		t.rows = NewNodes()
	}
	if t.Selected == nil {
		t.Selected = NewString("")
	}
}

func (t *TargetHour) Init(_ Ctx) { t.ensure() }

func (t *TargetHour) SetSelectMode(on bool)        { t.sel.SetOn(on) }
func (t *TargetHour) SetDanger(on bool)            { t.sel.SetDanger(on) }
func (t *TargetHour) OnCheckedChange(fn func(int)) { t.sel.OnChange = fn }

// itemIDs is the "current rows" listselect.Header/RowOf read to size the "k /
// N" count and the select-all tri-state. t.items is a plain field, not a
// signal — reading t.rows.Get() first is what makes a derive that calls
// itemIDs() re-run on every SetItems (a reload, a filter, a day switch), not
// just on a selection change. Skipping this read is the exact bug that left
// the header's count frozen after a reload: the derive had nothing here to
// resubscribe to.
func (t *TargetHour) itemIDs() []string {
	_ = t.rows.Get()
	ids := make([]string, len(t.items))
	for i, it := range t.items {
		ids[i] = it.ID
	}
	return ids
}

func (t *TargetHour) CheckedIDs() []string {
	return t.sel.CheckedIDs(t.itemIDs())
}

func (t *TargetHour) SetItems(items []Item) {
	t.ensure()
	t.items = items
	nodes := make([]*Element, 0, len(items)+len(t.FreeSlots))
	for _, it := range items {
		nodes = append(nodes, t.buildRow(it))
	}
	for _, hhmm := range t.FreeSlots {
		nodes = append(nodes, t.buildFreeSlot(hhmm))
	}
	t.rows.Set(nodes)
}

func (t *TargetHour) Items() []Item { return t.items }
func (t *TargetHour) Count() int    { return len(t.items) }

func (t *TargetHour) Render() *Element {
	list := Ul().Set(clsList.AsAttr()).Attr("role", "listbox").BindChildren(t.rows)
	return Div().Set(clsListWrap.AsAttr()).
		BindStateFunc(widget.Open, func() bool { return t.sel.On().Get() }).
		Child(listselect.Header(&t.sel, t.itemIDs, t.WidgetName())).
		Child(list)
}

func (t *TargetHour) buildRow(it Item) *Element {
	id := it.ID
	key := "th-" + id

	// RowOf owns the per-row selection wiring: the narrow Edit/Danger derives
	// and the check box. isSel widens Edit with the normal-mode highlight for
	// the ROW only; the box binds RowOf's narrow ones, so a row merely loaded
	// in normal mode reveals no glyph.
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

	st := StatusPending
	if t.StatusOf != nil {
		st = t.StatusOf(it)
	}
	row.BindStateFunc(widget.Locked, func() bool { return st == StatusConfirmed })
	row.BindStateFunc(widget.Busy, func() bool { return st == StatusAttended })

	row.On("click", func(Event) {
		if t.sel.On().Get() {
			t.sel.Toggle(id)
			return
		}
		if t.OnSelect != nil {
			t.OnSelect(it)
		}
	})

	hour := Span().Set(clsHour.AsAttr()).Text(it.LeadMain)

	content := Div().Set(clsContent.AsAttr()).
		Child(hour).
		Child(r.Check).
		Child(Span().Set(clsLabel.AsAttr()).Text(it.Label))
	if it.Description != "" {
		content.Child(Span().Set(clsBadge.AsAttr()).
			Attr("title", it.Description).
			Text(fmt.Convert(it.Description).Truncate(badgeChars).String()))
	}

	row.Child(content)
	return row
}

// buildFreeSlot renders one reservable free hour as a light row distinct from
// a real Item: no selection chrome, no tint — it is an action ("reserve this
// hour"), not a record. Clicking invokes OnPickFree.
func (t *TargetHour) buildFreeSlot(hhmm string) *Element {
	key := "th-free-" + hhmm

	content := Div().Set(clsContent.AsAttr()).
		Child(Span().Set(clsHour.AsAttr()).Text(hhmm)).
		Child(Span().Set(clsLabel.AsAttr()).Text("")).
		Child(Span().Set(clsFreeAdd.AsAttr()).Text("+"))

	return Li().Set(clsFree.AsAttr()).
		Key(key).
		Attr("data-free", hhmm).
		On("click", func(Event) {
			if t.sel.On().Get() {
				return // selection mode never opens on a free slot
			}
			if t.OnPickFree != nil {
				t.OnPickFree(hhmm)
			}
		}).
		Child(content)
}
