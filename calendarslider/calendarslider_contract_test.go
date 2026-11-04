package calendarslider

import (
	"testing"
	"webtyp.com/time"

	"webtyp.com/dom"
	"webtyp.com/widget"
)

var _ dom.Component = (*CalendarSlider)(nil)

func TestCalendarSlider_InitTwiceSafe(t *testing.T) {
	c := &CalendarSlider{Zone: time.MachineZone()}
	c.Init(nil)
	c.Init(nil)
}

func TestCalendarSlider_RenderIdempotent(t *testing.T) {
	c := &CalendarSlider{Zone: time.MachineZone(), Start: "2026-08"}
	c.Init(nil)
	c.today = "2026-08-11"

	first := c.Render().String()
	second := c.Render().String()

	if first != second {
		t.Errorf("Render() not idempotent: first=%q second=%q", first, second)
	}
}

func TestCalendarSlider_KindAllowsSelected(t *testing.T) {
	if !(&CalendarSlider{Zone: time.MachineZone()}).WidgetKind().Allows(widget.Selected) {
		t.Error("kind Grid should allow the Selected state used by When(Selected, PartDay)")
	}
}
