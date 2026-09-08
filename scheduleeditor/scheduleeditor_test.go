//go:build !wasm

package scheduleeditor

import (
	"strings"
	"testing"
)

type emptyCtx struct{}

func (emptyCtx) OnCleanup(func()) {}

func testEditor() *ScheduleEditor {
	return &ScheduleEditor{
		Pattern: []PatternRow{
			{StartMin: 540, EndMin: 1080, Days: []int{1, 3, 5}},
			{StartMin: 900, EndMin: 1140, Days: []int{1, 3, 5}},
		},
		Marked: []MarkedDay{
			{Date: "2026-09-19", StartMin: 540, EndMin: 780},
		},
		Bounds: Bounds{OpenMin: 480, CloseMin: 1200},
	}
}

func TestOnePatternRowCoversSeveralWeekdays(t *testing.T) {
	e := &ScheduleEditor{
		Pattern: []PatternRow{
			{StartMin: 540, EndMin: 1080, Days: []int{1, 3, 5}},
		},
	}
	e.Init(&emptyCtx{})
	html := e.Render().String()

	if !strings.Contains(html, "scheduleeditor__pattern-row") {
		t.Fatalf("expected pattern row in markup:\n%s", html)
	}
	if strings.Count(html, "checked") < 3 {
		t.Errorf("expected at least 3 checked day chips for Mon/Wed/Fri:\n%s", html)
	}
}

func TestTwoRowsShareADayAndLeaveAGap(t *testing.T) {
	e := &ScheduleEditor{
		Pattern: []PatternRow{
			{StartMin: 540, EndMin: 780, Days: []int{1}},
			{StartMin: 900, EndMin: 1140, Days: []int{1}},
		},
	}
	e.Init(&emptyCtx{})
	html := e.Render().String()

	if strings.Contains(html, "data-invalid='true'") {
		t.Errorf("two non-overlapping rows on same day should be valid:\n%s", html)
	}
	if strings.Count(html, "scheduleeditor__pattern-row") != 2 {
		t.Errorf("expected 2 pattern rows rendered:\n%s", html)
	}
}

func TestEmptyPatternWithMarkedDaysIsValid(t *testing.T) {
	e := &ScheduleEditor{
		Pattern: []PatternRow{},
		Marked: []MarkedDay{
			{Date: "2026-09-19", StartMin: 540, EndMin: 780},
		},
	}
	e.Init(&emptyCtx{})
	html := e.Render().String()

	if strings.Contains(html, "data-invalid='true'") {
		t.Errorf("empty pattern with marked days should be valid:\n%s", html)
	}
	if !strings.Contains(html, "2026-09-19") {
		t.Errorf("marked day date should be rendered:\n%s", html)
	}
}

func TestMarkingDaysUsesTheCommonWindow(t *testing.T) {
	e := &ScheduleEditor{}
	e.Init(&emptyCtx{})
	e.markerStart.Set("480")
	e.markerEnd.Set("960")

	var gotDates []string
	var gotStart, gotEnd int
	e.OnDaysMarked = func(dates []string, startMin, endMin int) {
		gotDates = dates
		gotStart = startMin
		gotEnd = endMin
	}

	e.handleDayToggle("2026-09-20", true)

	if len(gotDates) != 1 || gotDates[0] != "2026-09-20" {
		t.Fatalf("OnDaysMarked received %v, want ['2026-09-20']", gotDates)
	}
	if gotStart != 480 || gotEnd != 960 {
		t.Errorf("OnDaysMarked hours = (%d, %d), want (480, 960)", gotStart, gotEnd)
	}
}

func TestMarkedDayCanDivergeFromTheCommonWindow(t *testing.T) {
	e := &ScheduleEditor{
		Marked: []MarkedDay{
			{Date: "2026-09-19", StartMin: 480, EndMin: 720},
		},
	}
	e.Init(&emptyCtx{})
	html := e.Render().String()

	if !strings.Contains(html, "2026-09-19") {
		t.Fatalf("marked day must render date in list:\n%s", html)
	}
	if !strings.Contains(html, "value='480' selected=''") || !strings.Contains(html, "value='720' selected=''") {
		t.Errorf("divergent hours 08:00/12:00 should be selected in marked day item:\n%s", html)
	}
}

func TestPatternAndMarkedDaysRenderTogether(t *testing.T) {
	e := testEditor()
	e.Init(&emptyCtx{})
	html := e.Render().String()

	if !strings.Contains(html, "scheduleeditor__pattern") {
		t.Errorf("pattern section missing:\n%s", html)
	}
	if !strings.Contains(html, "scheduleeditor__marker") {
		t.Errorf("marker section missing:\n%s", html)
	}
}

func TestUnmarkingADayFiresOnDaysUnmarked(t *testing.T) {
	e := &ScheduleEditor{}
	e.Init(&emptyCtx{})

	var gotUnmarked []string
	e.OnDaysUnmarked = func(dates []string) {
		gotUnmarked = dates
	}

	e.handleDayToggle("2026-09-20", false)

	if len(gotUnmarked) != 1 || gotUnmarked[0] != "2026-09-20" {
		t.Fatalf("OnDaysUnmarked received %v, want ['2026-09-20']", gotUnmarked)
	}
}

func TestHourOptionsAreClampedToBounds(t *testing.T) {
	b := Bounds{OpenMin: 480, CloseMin: 1080}
	opts := hourOptions(540, b, 15)

	if len(opts) != 41 {
		t.Fatalf("expected 41 options for 08:00..18:00 step 15, got %d", len(opts))
	}
	if !strings.Contains(opts[0].String(), "08:00") {
		t.Errorf("first option = %s, want 08:00", opts[0].String())
	}
	if !strings.Contains(opts[len(opts)-1].String(), "18:00") {
		t.Errorf("last option = %s, want 18:00", opts[len(opts)-1].String())
	}
}

func TestWiderBoundsOfferMoreOptions(t *testing.T) {
	narrow := hourOptions(540, Bounds{OpenMin: 540, CloseMin: 1020}, 15)
	wider := hourOptions(540, Bounds{OpenMin: 480, CloseMin: 1200}, 15)

	if len(wider) <= len(narrow) {
		t.Fatalf("wider bounds should produce more options: wider=%d, narrow=%d", len(wider), len(narrow))
	}
}

func TestZeroBoundsFallBackToFullDay(t *testing.T) {
	opts := hourOptions(0, Bounds{0, 0}, 15)
	if len(opts) != 96 {
		t.Fatalf("zero bounds should fallback to 96 options (00:00..23:45), got %d", len(opts))
	}
	if !strings.Contains(opts[0].String(), "00:00") {
		t.Errorf("first option = %s, want 00:00", opts[0].String())
	}
	if !strings.Contains(opts[len(opts)-1].String(), "23:45") {
		t.Errorf("last option = %s, want 23:45", opts[len(opts)-1].String())
	}
}

func TestHolidayIsNotSelectableInTheMarker(t *testing.T) {
	e := &ScheduleEditor{
		Holidays: []string{"2026-09-18"},
	}
	e.Init(&emptyCtx{})
	html := e.Render().String()

	if strings.Contains(html, "data-date='2026-09-18' class='calendarslider__day calendarslider__day-selectable") {
		t.Errorf("holiday should not be marked selectable in calendar:\n%s", html)
	}
}

func TestOverlappingRowsAreMarkedInvalidButNotBlocked(t *testing.T) {
	e := &ScheduleEditor{
		Pattern: []PatternRow{
			{StartMin: 540, EndMin: 1080, Days: []int{1}},
			{StartMin: 600, EndMin: 900, Days: []int{1}},
		},
	}
	e.Init(&emptyCtx{})
	html := e.Render().String()

	if !strings.Contains(html, "data-invalid='true'") {
		t.Errorf("overlapping pattern rows must carry data-invalid='true':\n%s", html)
	}
	if strings.Count(html, "scheduleeditor__pattern-row") != 2 {
		t.Errorf("overlapping pattern rows should still render both rows:\n%s", html)
	}
}

func TestRevealedStatesAreWrittenByTheMarkup(t *testing.T) {
	e := &ScheduleEditor{
		Pattern: []PatternRow{
			{StartMin: 1080, EndMin: 540, Days: []int{1}},
		},
	}
	e.Init(&emptyCtx{})
	e.sel.Set("2026-09-19")
	html := e.Render().String()

	for _, kv := range e.sheet().StateAttrs() {
		if !strings.Contains(html, kv.Key) {
			t.Errorf("stylesheet reveals/repaints on %q (value: %q) but no element writes it:\n%s", kv.Key, kv.Value, html)
		}
	}
}

func TestExceptions_ListSorted(t *testing.T) {
	e := testEditor()
	e.Exceptions = []Exception{
		{ID: "b", Date: "2026-09-20", Type: ExcBlocked},
		{ID: "a", Date: "2026-09-18", Type: ExcHoliday},
		{ID: "c", Date: "2026-09-19", Type: ExcSpecialHours, StartMin: 540, EndMin: 600},
	}
	e.Init(&emptyCtx{})
	list := e.buildExceptionList()

	items := []string{}
	for _, c := range list.Children() {
		s := extractDate(c.String())
		if s != "" {
			items = append(items, s)
		}
	}
	if len(items) != 3 || items[0] != "2026-09-18" || items[1] != "2026-09-19" || items[2] != "2026-09-20" {
		t.Fatalf("exceptions not sorted by date: %v", items)
	}
}

func extractDate(s string) string {
	quote := strings.Index(s, "2026-09-")
	if quote < 0 {
		return ""
	}
	return s[quote : quote+10]
}

func TestExceptions_HolidayReadonly(t *testing.T) {
	e := testEditor()
	e.Holidays = []string{"2026-09-18"}
	e.Exceptions = []Exception{
		{ID: "a", Date: "2026-09-18", Type: ExcHoliday},
	}
	e.Init(&emptyCtx{})

	html := e.buildExceptionList().String()
	if strings.Contains(html, "scheduleeditor__exc-remove") {
		t.Errorf("holiday exception must not render a Remove button:\n%s", html)
	}
	if !strings.Contains(html, "scheduleeditor__exc-holiday") {
		t.Errorf("holiday-exception row must carry the holiday mark:\n%s", html)
	}
}
