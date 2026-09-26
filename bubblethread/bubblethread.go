package bubblethread

import (
	. "webtyp.com/dom"
	. "webtyp.com/html"
	"webtyp.com/widget"
)

// NameBubbleThread is the widget identity.
const NameBubbleThread = widget.Name("bubblethread")

const (
	PartThread = widget.Part("thread")
	PartMine   = widget.Part("mine")
	PartTheirs = widget.Part("theirs")
	PartAuthor = widget.Part("author")
	PartBody   = widget.Part("body")
	PartTime   = widget.Part("time")
	PartRead   = widget.Part("read")
	PartEmpty  = widget.Part("empty")
)

var (
	clsThreadWrap = NameBubbleThread.Root()
	clsThread     = NameBubbleThread.Class(PartThread)
	clsMine       = NameBubbleThread.Class(PartMine)
	clsTheirs     = NameBubbleThread.Class(PartTheirs)
	clsAuthor     = NameBubbleThread.Class(PartAuthor)
	clsBody       = NameBubbleThread.Class(PartBody)
	clsTime       = NameBubbleThread.Class(PartTime)
	clsRead       = NameBubbleThread.Class(PartRead)
	clsEmpty      = NameBubbleThread.Class(PartEmpty)
)

type Bubble struct {
	ID     string
	Author string // shown only when Mine is false
	Body   string
	Time   string
	Mine   bool
	Read   bool // meaningful only when Mine is true
}

type BubbleThread struct {
	Element
	ReadLabel string // text under a Mine && Read bubble, e.g. "Leído"; "" renders nothing
	Empty     string // text shown with no bubbles; "" renders nothing

	bubbles []Bubble
	nodes   *SignalNodes
}

func (t *BubbleThread) WidgetName() widget.Name { return NameBubbleThread }
func (t *BubbleThread) WidgetKind() widget.Kind { return widget.Region }

func (t *BubbleThread) ensure() {
	if t.nodes == nil {
		t.nodes = NewNodes()
	}
}

func (t *BubbleThread) Init(_ Ctx) { t.ensure() }

func (t *BubbleThread) SetBubbles(bubbles []Bubble) {
	t.ensure()
	t.bubbles = bubbles
	t.rebuildNodes()
}

func (t *BubbleThread) Append(bubbles ...Bubble) {
	t.ensure()
	added := false
	for _, b := range bubbles {
		if t.hasBubble(b.ID) {
			continue
		}
		t.bubbles = append(t.bubbles, b)
		added = true
	}
	if added {
		t.rebuildNodes()
	}
}

func (t *BubbleThread) MarkRead(ids ...string) {
	t.ensure()
	modified := false
	for _, id := range ids {
		for i := range t.bubbles {
			if t.bubbles[i].ID == id {
				if t.bubbles[i].Mine && !t.bubbles[i].Read {
					t.bubbles[i].Read = true
					modified = true
				}
			}
		}
	}
	if modified {
		t.rebuildNodes()
	}
}

func (t *BubbleThread) Bubbles() []Bubble { return t.bubbles }

func (t *BubbleThread) hasBubble(id string) bool {
	for _, b := range t.bubbles {
		if b.ID == id {
			return true
		}
	}
	return false
}

func (t *BubbleThread) rebuildNodes() {
	nodes := make([]*Element, 0, len(t.bubbles))
	for _, b := range t.bubbles {
		nodes = append(nodes, t.buildBubble(b))
	}
	t.nodes.Set(nodes)
	if len(nodes) > 0 {
		scrollToLast(nodes[len(nodes)-1].GetID())
	}
}

func (t *BubbleThread) Render() *Element {
	t.ensure()

	thread := Div().Set(clsThread.AsAttr()).
		Attr("role", "log").
		Attr("aria-live", "polite").
		BindChildren(t.nodes)

	emptyShow := Show(DeriveBool(func() bool {
		_ = t.nodes.Get()
		return len(t.bubbles) == 0 && t.Empty != ""
	}), Div().Set(clsEmpty.AsAttr()).Text(t.Empty))

	return Div().Set(clsThreadWrap.AsAttr()).
		Child(thread).
		Child(emptyShow)
}

func (t *BubbleThread) buildBubble(b Bubble) *Element {
	key := "bt-" + b.ID

	cls := clsTheirs
	if b.Mine {
		cls = clsMine
	}

	el := Div().Set(cls.AsAttr()).
		Key(key).
		Attr("data-id", b.ID)

	if !b.Mine && b.Author != "" {
		el.Child(Div().Set(clsAuthor.AsAttr()).Text(b.Author))
	}

	el.Child(Div().Set(clsBody.AsAttr()).Text(b.Body))
	el.Child(Div().Set(clsTime.AsAttr()).Text(b.Time))

	if b.Mine && b.Read && t.ReadLabel != "" {
		el.Child(Div().Set(clsRead.AsAttr()).Text(t.ReadLabel))
	}

	return el
}
