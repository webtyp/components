package presencelist

import (
	. "webtyp.com/dom"
	. "webtyp.com/html"
	"webtyp.com/widget"
)

// NamePresenceList is the widget identity.
const NamePresenceList = widget.Name("presencelist")

const (
	PartList       = widget.Part("list")
	PartRow        = widget.Part("row")
	PartDotOnline  = widget.Part("dot-online")
	PartDotOffline = widget.Part("dot-offline")
	PartLabel      = widget.Part("label")
	PartStatus     = widget.Part("status")
	PartEmpty      = widget.Part("empty")
)

var (
	clsListWrap   = NamePresenceList.Root()
	clsList       = NamePresenceList.Class(PartList)
	clsRow        = NamePresenceList.Class(PartRow)
	clsDotOnline  = NamePresenceList.Class(PartDotOnline)
	clsDotOffline = NamePresenceList.Class(PartDotOffline)
	clsLabel      = NamePresenceList.Class(PartLabel)
	clsStatus     = NamePresenceList.Class(PartStatus)
	clsEmpty      = NamePresenceList.Class(PartEmpty)
)

type Person struct {
	ID     string
	Label  string
	Online bool
}

type PresenceList struct {
	Element
	OnSelect     func(id string) // row clicked
	OnlineLabel  string          // screen-reader text for the online dot, e.g. "En línea"
	OfflineLabel string          // e.g. "Desconectado"
	Empty        string

	people []Person
	rows   *SignalNodes
}

func (l *PresenceList) WidgetName() widget.Name { return NamePresenceList }
func (l *PresenceList) WidgetKind() widget.Kind { return widget.Listbox }

func (l *PresenceList) ensure() {
	if l.rows == nil {
		l.rows = NewNodes()
	}
}

func (l *PresenceList) Init(_ Ctx) { l.ensure() }

func (l *PresenceList) SetPeople(p []Person) {
	l.ensure()
	sorted := sortPeople(p)
	l.people = sorted

	nodes := make([]*Element, 0, len(sorted))
	for _, person := range sorted {
		nodes = append(nodes, l.buildRow(person))
	}
	l.rows.Set(nodes)
}

func (l *PresenceList) People() []Person { return l.people }

func sortPeople(p []Person) []Person {
	out := make([]Person, len(p))
	copy(out, p)
	for i := 1; i < len(out); i++ {
		key := out[i]
		j := i - 1
		for j >= 0 && personLess(key, out[j]) {
			out[j+1] = out[j]
			j--
		}
		out[j+1] = key
	}
	return out
}

func personLess(a, b Person) bool {
	if a.Online != b.Online {
		return a.Online
	}
	return a.Label < b.Label
}

func (l *PresenceList) Render() *Element {
	l.ensure()

	list := Ul().Set(clsList.AsAttr()).Attr("role", "listbox").BindChildren(l.rows)

	emptyShow := Show(DeriveBool(func() bool {
		_ = l.rows.Get()
		return len(l.people) == 0 && l.Empty != ""
	}), Div().Set(clsEmpty.AsAttr()).Text(l.Empty))

	return Div().Set(clsListWrap.AsAttr()).
		Child(list).
		Child(emptyShow)
}

func (l *PresenceList) buildRow(p Person) *Element {
	key := "pl-" + p.ID
	id := p.ID

	row := Button().Set(clsRow.AsAttr()).
		Key(key).
		Attr("type", "button").
		Attr("role", "option")

	row.OnClick(func(Event) {
		if l.OnSelect != nil {
			l.OnSelect(id)
		}
	})

	dotCls := clsDotOffline
	statusText := l.OfflineLabel
	if p.Online {
		dotCls = clsDotOnline
		statusText = l.OnlineLabel
	}

	row.Child(Span().Set(dotCls.AsAttr()))
	row.Child(Span().Set(clsLabel.AsAttr()).Text(p.Label))
	if statusText != "" {
		row.Child(Span().Set(clsStatus.AsAttr()).Text(statusText))
	}

	return row
}
