package inboxlist

import (
	"webtyp.com/components/countbadge"
	. "webtyp.com/dom"
	"webtyp.com/fmt"
	. "webtyp.com/html"
	"webtyp.com/widget"
)

// NameInboxList is the widget identity.
const NameInboxList = widget.Name("inboxlist")

const (
	PartList        = widget.Part("list")
	PartRow         = widget.Part("row")
	PartTitle       = widget.Part("title")
	PartTitleUnread = widget.Part("title-unread")
	PartTime        = widget.Part("time")
	PartPreview     = widget.Part("preview")
	PartEmpty       = widget.Part("empty")
)

var (
	clsListWrap    = NameInboxList.Root()
	clsList        = NameInboxList.Class(PartList)
	clsRow         = NameInboxList.Class(PartRow)
	clsTitle       = NameInboxList.Class(PartTitle)
	clsTitleUnread = NameInboxList.Class(PartTitleUnread)
	clsTime        = NameInboxList.Class(PartTime)
	clsPreview     = NameInboxList.Class(PartPreview)
	clsEmpty       = NameInboxList.Class(PartEmpty)
)

type Row struct {
	ID      string
	Title   string
	Preview string // already short; rendered as-is
	Time    string // "HH:MM" or a short date; rendered as-is
	Unread  int
}

type InboxList struct {
	Element
	Selected *SignalString   // optional; created when nil. Holds the selected Row.ID.
	OnSelect func(id string) // called when a row is clicked; also sets Selected
	Empty    string          // text shown when there are no rows; "" shows nothing

	items []Row
	rows  *SignalNodes
}

func (l *InboxList) WidgetName() widget.Name { return NameInboxList }
func (l *InboxList) WidgetKind() widget.Kind { return widget.Listbox }

func (l *InboxList) ensure() {
	if l.rows == nil {
		l.rows = NewNodes()
	}
	if l.Selected == nil {
		l.Selected = NewString("")
	}
}

func (l *InboxList) Init(_ Ctx) { l.ensure() }

func (l *InboxList) SetRows(rows []Row) {
	l.ensure()
	l.items = rows

	nodes := make([]*Element, 0, len(rows))
	for _, row := range rows {
		r := row
		nodes = append(nodes, l.buildRow(r))
	}
	l.rows.Set(nodes)
}

func (l *InboxList) Rows() []Row { return l.items }

func (l *InboxList) Render() *Element {
	l.ensure()

	list := Ul().Set(clsList.AsAttr()).Attr("role", "listbox").BindChildren(l.rows)

	emptyShow := Show(DeriveBool(func() bool {
		_ = l.rows.Get()
		return len(l.items) == 0 && l.Empty != ""
	}), Div().Set(clsEmpty.AsAttr()).Text(l.Empty))

	return Div().Set(clsListWrap.AsAttr()).
		Child(list).
		Child(emptyShow)
}

func (l *InboxList) buildRow(row Row) *Element {
	id := row.ID
	key := "inbox-" + id

	titleCls := clsTitle
	if row.Unread > 0 {
		titleCls = clsTitleUnread
	}

	btn := Button().Set(clsRow.AsAttr()).
		Key(key).
		Attr("type", "button").
		Attr("role", "option").
		BindStateFunc(widget.Selected, func() bool {
			return l.Selected.Get() == id
		}).
		BindAttrBool("aria-selected", DeriveBool(func() bool {
			return l.Selected.Get() == id
		}))

	btn.OnClick(func(Event) {
		if l.Selected != nil {
			l.Selected.Set(id)
		}
		if l.OnSelect != nil {
			l.OnSelect(id)
		}
	})

	badge := &countbadge.CountBadge{
		Count:   NewString(fmt.Sprint(row.Unread)),
		Visible: NewBool(row.Unread > 0),
	}

	btn.Child(
		Div().Child(Span().Set(titleCls.AsAttr()).Text(row.Title)).Child(Span().Set(clsTime.AsAttr()).Text(row.Time)),
	).Child(
		Div().Set(clsPreview.AsAttr()).Text(row.Preview),
	).Child(
		badge.Render(),
	)

	return btn
}
