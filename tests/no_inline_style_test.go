package components_test

import (
	"strings"
	"testing"

	"webtyp.com/components/calendarslider"
	"webtyp.com/components/cellgrid"
	"webtyp.com/time"
)

func TestComponents_NoInlineStyleAttr(t *testing.T) {
	cal := &calendarslider.CalendarSlider{
		Zone:  time.MachineZone(),
		Start: "2026-08",
		Occupation: []calendarslider.OccupationDay{
			{Date: "2026-08-15", Percent: 50},
		},
	}
	cal.Init(nil)
	calHTML := cal.Render().String()
	if strings.Contains(calHTML, "style=") {
		t.Fatalf("CalendarSlider must not emit inline style= attribute, got: %s", calHTML)
	}

	grid := &cellgrid.CellGrid{Rows: 2, Cols: 3}
	gridHTML := grid.Render().String()
	if strings.Contains(gridHTML, "style=") {
		t.Fatalf("CellGrid must not emit inline style= attribute, got: %s", gridHTML)
	}
}
