package stepaccordion

import (
	"webtyp.com/dom"
	"webtyp.com/lang"
	"webtyp.com/widget"
)

// NameStepAccordion is the widget name for stepaccordion.
const NameStepAccordion = widget.Name("stepaccordion")

const (
	PartStep    = widget.Part("step")
	PartHeader  = widget.Part("header")
	PartBadge   = widget.Part("badge")
	PartTitle   = widget.Part("title")
	PartSummary = widget.Part("summary")
	PartEdit    = widget.Part("edit")
	PartBody    = widget.Part("body")
)

// Step is one section of the wizard.
type Step struct {
	Key   string
	Title lang.Text
	Body  dom.Component
}

// StepAccordion opens its steps one at a time, in order.
type StepAccordion struct {
	dom.Element
	Steps    []Step
	OnReopen func(key string)
}

func (a *StepAccordion) WidgetName() widget.Name { return NameStepAccordion }
func (a *StepAccordion) WidgetKind() widget.Kind { return widget.Disclosure }

func (a *StepAccordion) Init(_ dom.Ctx)         { panic("not implemented") }
func (a *StepAccordion) Render() *dom.Element   { panic("not implemented") }
func (a *StepAccordion) Advance(summary string) { panic("not implemented") }
func (a *StepAccordion) Reopen(key string)      { panic("not implemented") }
func (a *StepAccordion) Open() string           { panic("not implemented") }
