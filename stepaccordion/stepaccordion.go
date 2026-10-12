package stepaccordion

import (
	"webtyp.com/dom"
	"webtyp.com/fmt"
	"webtyp.com/html"
	"webtyp.com/lang"
	"webtyp.com/widget"
)

// NameStepAccordion is the widget name for stepaccordion.
const NameStepAccordion = widget.Name("stepaccordion")

const (
	PartStep    = widget.Part("step")    // <section> per step
	PartHeader  = widget.Part("header")  // number badge + title + summary + edit button
	PartBadge   = widget.Part("badge")   // step number "1", "2", …
	PartTitle   = widget.Part("title")   // the step heading
	PartSummary = widget.Part("summary") // one line, shown only on a finished step
	PartEdit    = widget.Part("edit")    // "Edit" button, shown only on a finished step
	PartBody    = widget.Part("body")    // the step content, shown only on the open step
)

var (
	clsRoot    = NameStepAccordion.Root()
	clsStep    = NameStepAccordion.Class(PartStep)
	clsHeader  = NameStepAccordion.Class(PartHeader)
	clsBadge   = NameStepAccordion.Class(PartBadge)
	clsTitle   = NameStepAccordion.Class(PartTitle)
	clsSummary = NameStepAccordion.Class(PartSummary)
	clsEdit    = NameStepAccordion.Class(PartEdit)
	clsBody    = NameStepAccordion.Class(PartBody)
)

// textEdit is the label of the button that reopens a finished step.
const textEdit lang.Text = "Edit"

// Step is one section of the wizard.
type Step struct {
	Key   string        // stable, unique within the accordion
	Title lang.Text     // English; translated in Render
	Body  dom.Component // rendered inside PartBody; shown while the step is open
}

// StepAccordion opens its steps one at a time, in order. A finished step
// collapses to a one-line summary with an "Edit" button; a step below the open
// one stays locked until every step above it is finished.
//
// Each step is in exactly one phase, written on its <section> as a state:
// open (widget.Open), done (widget.Locked: read-only but legible) or locked
// (widget.Disabled). The stylesheet reveals body, summary and edit from it.
type StepAccordion struct {
	dom.Element // value embed — NEVER pointer (TinyGo heap constraint)

	Steps []Step
	// OnReopen runs after Reopen(key), so the host can clear the choices that
	// depended on that step. Optional.
	OnReopen func(key string)

	phases []stepPhase // one per step, same order; a slice, never a map
	ready  bool
}

// stepPhase holds the signals of one step. locked is derived: neither open nor done.
type stepPhase struct {
	open    *dom.SignalBool
	done    *dom.SignalBool
	locked  *dom.SignalBool
	summary *dom.SignalString
}

func (a *StepAccordion) WidgetName() widget.Name { return NameStepAccordion }
func (a *StepAccordion) WidgetKind() widget.Kind { return widget.Disclosure }

// Init opens the first step and locks the rest. Calling it again is a no-op.
func (a *StepAccordion) Init(_ dom.Ctx) {
	if a.ready {
		return
	}
	a.ready = true
	a.phases = make([]stepPhase, len(a.Steps))
	for i := range a.Steps {
		p := stepPhase{
			open:    dom.NewBool(i == 0),
			done:    dom.NewBool(false),
			summary: dom.NewString(""),
		}
		p.locked = dom.DeriveBool(func() bool { return !p.open.Get() && !p.done.Get() })
		a.phases[i] = p
	}
}

func (a *StepAccordion) Render() *dom.Element {
	a.Init(nil)
	root := html.Div().Set(clsRoot.AsAttr())
	for i, st := range a.Steps {
		p := a.phases[i]
		key := st.Key

		title := html.H3().Set(clsTitle.AsAttr()).Text(lang.Translate(st.Title).String())
		body := html.Div().Set(clsBody.AsAttr()).
			Attr("role", "region").
			Attr("aria-labelledby", title.GetID())
		if st.Body != nil {
			body.Child(st.Body)
		}
		edit := html.Button().Set(clsEdit.AsAttr()).
			Attr("type", "button").
			Attr("aria-expanded", "false").
			Attr("aria-controls", body.GetID()).
			Text(lang.Translate(textEdit).String()).
			OnClick(func(dom.Event) { a.Reopen(key) })

		header := html.Div().Set(clsHeader.AsAttr()).Child(
			html.Span().Set(clsBadge.AsAttr()).Text(fmt.Convert(i+1).String()),
			title,
			html.Span().Set(clsSummary.AsAttr()).BindText(p.summary),
			edit,
		)

		root.Child(html.Section().Set(clsStep.AsAttr()).
			Key(key).
			BindState(widget.Open, p.open).
			BindState(widget.Locked, p.done).
			BindState(widget.Disabled, p.locked).
			Child(header, body))
	}
	return root
}

// Advance marks the open step finished, shows summary on its header, and opens
// the next step. On the last step it marks it finished and opens nothing. With
// no open step it does nothing.
func (a *StepAccordion) Advance(summary string) {
	a.Init(nil)
	i := a.openIndex()
	if i < 0 {
		return
	}
	a.phases[i].summary.Set(summary)
	a.phases[i].open.Set(false)
	a.phases[i].done.Set(true)
	if i+1 < len(a.phases) {
		a.phases[i+1].open.Set(true)
	}
}

// Reopen opens the finished step key again. Every step after it goes back to
// locked and loses its summary. Then OnReopen(key) runs. Unknown key or a step
// that is not finished: no-op.
func (a *StepAccordion) Reopen(key string) {
	a.Init(nil)
	at := -1
	for i, st := range a.Steps {
		if st.Key == key {
			at = i
			break
		}
	}
	if at < 0 || !a.phases[at].done.Get() {
		return
	}
	a.phases[at].done.Set(false)
	a.phases[at].summary.Set("")
	a.phases[at].open.Set(true)
	for i := at + 1; i < len(a.phases); i++ {
		a.phases[i].open.Set(false)
		a.phases[i].done.Set(false)
		a.phases[i].summary.Set("")
	}
	if a.OnReopen != nil {
		a.OnReopen(key)
	}
}

// Open returns the key of the open step, or "" when every step is finished.
func (a *StepAccordion) Open() string {
	a.Init(nil)
	if i := a.openIndex(); i >= 0 {
		return a.Steps[i].Key
	}
	return ""
}

func (a *StepAccordion) openIndex() int {
	for i, p := range a.phases {
		if p.open.Get() {
			return i
		}
	}
	return -1
}
