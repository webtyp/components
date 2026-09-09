// Package scheduleeditor is a pure component for editing a professional's
// weekly pattern plus per-date marked days and exceptions. It knows nothing
// about router, orm or appointment_booking: the host supplies the data and
// translates callbacks to persistence ops.
//
// Time shape = minutes from midnight (int), the same shape the
// appointment_booking module persists. Times are presented in HH:MM inside
// select options; the value carried is the integer.
package scheduleeditor

import (
	"webtyp.com/date"
	. "webtyp.com/dom"
	"webtyp.com/fmt"
	"webtyp.com/fmt/lang"
	. "webtyp.com/html"
	"webtyp.com/time"
	"webtyp.com/widget"

	"webtyp.com/components/calendarslider"
)

// NameScheduleEditor is the widget identity.
const NameScheduleEditor = widget.Name("scheduleeditor")

const (
	PartPattern     = widget.Part("pattern")
	PartPatternRow  = widget.Part("pattern-row")
	PartDayChips    = widget.Part("day-chips")
	PartDayChip     = widget.Part("day-chip")
	PartDayLabel    = widget.Part("day-label")
	PartFieldLabel  = widget.Part("field-label")
	PartRowRemove   = widget.Part("row-remove")
	PartRowAdd      = widget.Part("row-add")
	PartMarker      = widget.Part("marker")
	PartMarkerHours = widget.Part("marker-hours")
	PartExceptions  = widget.Part("exceptions")
	PartSlider      = widget.Part("slider")
	PartExcForm     = widget.Part("exc-form")
	PartExcHours    = widget.Part("exc-hours")
	PartExcType     = widget.Part("exc-type")
	PartExcNotes    = widget.Part("exc-notes")
	PartExcAdd      = widget.Part("exc-add")
	PartExcList     = widget.Part("exc-list")
	PartExcItem     = widget.Part("exc-item")
	PartExcHoliday  = widget.Part("exc-holiday")
	PartExcRemove   = widget.Part("exc-remove")
)

var (
	clsRoot        = NameScheduleEditor.Root()
	clsPattern     = NameScheduleEditor.Class(PartPattern)
	clsPatternRow  = NameScheduleEditor.Class(PartPatternRow)
	clsDayChips    = NameScheduleEditor.Class(PartDayChips)
	clsDayChip     = NameScheduleEditor.Class(PartDayChip)
	clsDayLabel    = NameScheduleEditor.Class(PartDayLabel)
	clsFieldLabel  = NameScheduleEditor.Class(PartFieldLabel)
	clsRowRemove   = NameScheduleEditor.Class(PartRowRemove)
	clsRowAdd      = NameScheduleEditor.Class(PartRowAdd)
	clsMarker      = NameScheduleEditor.Class(PartMarker)
	clsMarkerHours = NameScheduleEditor.Class(PartMarkerHours)
	clsExceptions  = NameScheduleEditor.Class(PartExceptions)
	clsSlider      = NameScheduleEditor.Class(PartSlider)
	clsExcForm     = NameScheduleEditor.Class(PartExcForm)
	clsExcHours    = NameScheduleEditor.Class(PartExcHours)
	clsExcType     = NameScheduleEditor.Class(PartExcType)
	clsExcNotes    = NameScheduleEditor.Class(PartExcNotes)
	clsExcAdd      = NameScheduleEditor.Class(PartExcAdd)
	clsExcList     = NameScheduleEditor.Class(PartExcList)
	clsExcItem     = NameScheduleEditor.Class(PartExcItem)
	clsExcHoliday  = NameScheduleEditor.Class(PartExcHoliday)
	clsExcRemove   = NameScheduleEditor.Class(PartExcRemove)
)

// Tipos de excepción (valor de Exception.Type). Son las claves que
// appointment_booking espera en work_calendar_exception.exception_type.
const (
	ExcHoliday      = "HOLIDAY"
	ExcSpecialHours = "SPECIAL_HOURS"
	ExcBlocked      = "BLOCKED"
)

// PatternRow is a time range plus the weekdays it applies to. Several rows may
// cover the same weekday — "morning 09:00–13:00, afternoon 15:00–19:00" is two
// rows sharing a day, and the lunch break is the GAP between them.
//
// There is deliberately no break field: a break that is a field can describe
// exactly one interruption, and a gap describes any number.
type PatternRow struct {
	StartMin, EndMin int   // minutes from midnight
	Days             []int // 0=Sunday … 6=Saturday
}

// MarkedDay is a concrete date the professional works, independent of the
// weekly pattern. It is what lets an irregular schedule exist at all: a
// professional with no PatternRow and a list of MarkedDay is fully expressed.
type MarkedDay struct {
	Date             string // "YYYY-MM-DD"
	StartMin, EndMin int
}

// Bounds is the establishment's opening window — the only hours the editor may
// offer (CU-02). It is INPUT: the component renders within it and never
// fetches or validates it. Zero value (0,0) means unbounded, for a host that
// has no institutional calendar.
type Bounds struct {
	OpenMin, CloseMin int
}

// Exception es una excepción por fecha. Type ∈ {ExcHoliday, ExcSpecialHours,
// ExcBlocked}; StartMin/EndMin se usan por SPECIAL_HOURS/BLOCKED.
type Exception struct {
	ID               string // "" en alta (el host asigna al persistir)
	Date             string // "YYYY-MM-DD"
	Type             string
	StartMin, EndMin int
	Notes            string
}

// ScheduleEditor edita la agenda de un profesional: plantilla semanal de filas
// de patrones + marcado de días + panel de excepciones. Componente puro.
type ScheduleEditor struct {
	Element // value embed — NEVER pointer

	// Pattern is the weekly template. Initial state — the host persists on
	// each callback and re-mounts with fresh data.
	Pattern []PatternRow
	// Marked are the concrete dates worked, sorted ascending.
	Marked []MarkedDay
	// Bounds caps every hour control (CU-02). Widening it upstream is what
	// makes CU-03 visible here with no rebuild.
	Bounds Bounds
	// Horizon is how many months the marking calendar shows. 0 → 6.
	Horizon int
	// Holidays and Closures are read-only dates the editor paints as
	// unavailable; picking one is refused by the calendar, not by a message.
	Holidays []string
	Closures []string

	OnPatternChange func(rows []PatternRow)
	OnDaysMarked    func(dates []string, startMin, endMin int)
	OnDaysUnmarked  func(dates []string)
	OnMarkedDayEdit func(day MarkedDay)

	Exceptions        []Exception
	OnExceptionAdd    func(Exception)
	OnExceptionRemove func(id string)

	markerStart *SignalString // start min string for day marker common window
	markerEnd   *SignalString // end min string for day marker common window
	markedSel   *SignalString // multi-select signal for day marker calendar slider

	sel     *SignalString // día elegido por el calendario ("" = form oculto)
	excType *SignalString // tipo elegido en el formulario de alta
	excFrom *SignalString // hora "desde" del formulario (minutos)
	excTo   *SignalString // hora "hasta" del formulario (minutos)
	excNote *SignalString // notas del formulario
}

func (e *ScheduleEditor) WidgetName() widget.Name { return NameScheduleEditor }
func (e *ScheduleEditor) WidgetKind() widget.Kind { return widget.Form }

func (e *ScheduleEditor) Init(_ Ctx) {
	if e.markerStart == nil {
		e.markerStart = NewString("540") // 09:00
	}
	if e.markerEnd == nil {
		e.markerEnd = NewString("780") // 13:00
	}
	if e.markedSel == nil {
		e.markedSel = NewString(e.markedDatesString())
	}
	if e.sel == nil {
		e.sel = NewString("")
	}
	if e.excType == nil {
		e.excType = NewString(ExcSpecialHours)
	}
	if e.excFrom == nil {
		e.excFrom = NewString("540") // 09:00
	}
	if e.excTo == nil {
		e.excTo = NewString("1080") // 18:00
	}
	if e.excNote == nil {
		e.excNote = NewString("")
	}
}

func (e *ScheduleEditor) markedDatesString() string {
	res := ""
	for i, m := range e.Marked {
		if i > 0 {
			res += " "
		}
		res += m.Date
	}
	return res
}

// ---------------------------------------------------------------------------
// Horas — la opción única y su valor mostrado.
// ---------------------------------------------------------------------------

// hhmm formatea minutos como "HH:MM".
func hhmm(minutes int) string {
	if minutes < 0 || minutes > 1439 {
		return "--:--"
	}
	h := minutes / 60
	m := minutes % 60
	hh := pad2(h)
	mm := pad2(m)
	return hh + ":" + mm
}

func pad2(n int) string {
	if n < 10 {
		return "0" + fmt.Convert(n).String()
	}
	return fmt.Convert(n).String()
}

// hourOptions returns the <option> set for an hour select, clamped to the
// establishment's opening window. Hours outside it are NOT rendered disabled —
// they are not rendered at all: an option that cannot legally be chosen has no
// reason to exist in the list (CU-02).
//
// Bounds{} (0,0) means the host has no institutional calendar; fall back to
// the full day, 00:00–23:45.
func hourOptions(selected int, b Bounds, step int) []*Element {
	if step <= 0 {
		step = 15
	}
	start := b.OpenMin
	end := b.CloseMin
	if start == 0 && end == 0 {
		start = 0
		end = 1425 // 23:45
	}
	opts := make([]*Element, 0, (end-start)/step+1)
	for m := start; m <= end; m += step {
		if m == selected {
			opts = append(opts, SelectedOption(fmt.Convert(m).String(), hhmm(m)))
		} else {
			opts = append(opts, Option(fmt.Convert(m).String(), hhmm(m)))
		}
	}
	return opts
}

// ---------------------------------------------------------------------------
// Plantilla semanal (Pattern)
// ---------------------------------------------------------------------------

var shortWeekdayKeys = [7]string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}

func dayChipLabel(index int) string {
	if index < 0 || index > 6 {
		return ""
	}
	return lang.Translate(shortWeekdayKeys[index]).String()
}

func (e *ScheduleEditor) buildPattern() *Element {
	container := Div().Set(clsPattern.AsAttr())
	container.Child(Div().Text(lang.Translate("Weekly pattern").String()))

	for i, row := range e.Pattern {
		container.Child(e.buildPatternRow(i, row))
	}

	addBtn := Button().Set(clsRowAdd.AsAttr()).
		Attr("type", "button").
		Text(lang.Translate("Add row").String())
	addBtn.On("click", func(Event) {
		e.addPatternRow()
	})
	container.Child(addBtn)

	return container
}

func (e *ScheduleEditor) buildPatternRow(index int, row PatternRow) *Element {
	invalid := patternRowInvalid(index, row, e.Pattern)
	if invalid {
		Log("scheduleeditor: pattern row invalid or overlapping")
	}

	rowEl := Div().Set(clsPatternRow.AsAttr()).
		BindState(widget.Invalid, NewBool(invalid))

	// Start time select, with its label: two bare times side by side say
	// nothing about which one opens the block.
	rowEl.Child(Span().Set(clsFieldLabel.AsAttr()).Text(lang.Translate("From").String()))
	startSel := NewElement("select").Attr("name", "start-time")
	for _, opt := range hourOptions(row.StartMin, e.Bounds, 15) {
		startSel.Child(opt)
	}
	startSel.On("change", func(ev Event) {
		m, err := fmt.Convert(ev.TargetValue()).Int()
		if err == nil {
			e.updatePatternRow(index, func(r *PatternRow) { r.StartMin = m })
		}
	})
	rowEl.Child(startSel)

	// End time select
	rowEl.Child(Span().Set(clsFieldLabel.AsAttr()).Text(lang.Translate("To").String()))
	endSel := NewElement("select").Attr("name", "end-time")
	for _, opt := range hourOptions(row.EndMin, e.Bounds, 15) {
		endSel.Child(opt)
	}
	endSel.On("change", func(ev Event) {
		m, err := fmt.Convert(ev.TargetValue()).Int()
		if err == nil {
			e.updatePatternRow(index, func(r *PatternRow) { r.EndMin = m })
		}
	})
	rowEl.Child(endSel)

	// Day chips
	// Day chips in the app's own week order — Monday first unless the app
	// called date.SetFirstWeekday. The component never assumes a locale.
	chips := Div().Set(clsDayChips.AsAttr())
	for _, d := range date.WeekOrder() {
		dVal := d
		hasDay := containsInt(row.Days, dVal)

		chipInput := Input("checkbox").Set(clsDayChip.AsAttr()).
			Key("chip-"+fmt.Convert(index).String()+"-"+fmt.Convert(dVal).String()).
			BindAttrBool("checked", NewBool(hasDay))

		chipInput.On("change", func(ev Event) {
			checked := ev.TargetChecked()
			e.updatePatternRow(index, func(r *PatternRow) {
				if checked {
					if !containsInt(r.Days, dVal) {
						r.Days = append(r.Days, dVal)
					}
				} else {
					r.Days = removeInt(r.Days, dVal)
				}
			})
		})

		// The input stays the real control — VisuallyHidden keeps it focusable
		// and announced — and the label is the pill the eye sees, carrying the
		// chosen state. A native checkbox cannot be skinned; this pairing is
		// the standard way to make one read as a chip.
		label := Label().For(chipInput).
			Set(clsDayLabel.AsAttr()).
			BindState(widget.Selected, NewBool(hasDay)).
			Text(dayChipLabel(dVal))
		chips.Child(chipInput).Child(label)
	}
	rowEl.Child(chips)

	// Remove row button
	remBtn := Button().Set(clsRowRemove.AsAttr()).
		Attr("type", "button").
		Text(lang.Translate("Remove row").String())
	remBtn.On("click", func(Event) {
		e.removePatternRow(index)
	})
	rowEl.Child(remBtn)

	return rowEl
}

func (e *ScheduleEditor) addPatternRow() {
	newRow := PatternRow{
		StartMin: 540,                  // 09:00
		EndMin:   1080,                 // 18:00
		Days:     []int{1, 2, 3, 4, 5}, // Mon-Fri
	}
	newPattern := append(append([]PatternRow{}, e.Pattern...), newRow)
	if e.OnPatternChange != nil {
		e.OnPatternChange(newPattern)
	}
}

func (e *ScheduleEditor) updatePatternRow(index int, mutate func(*PatternRow)) {
	if index < 0 || index >= len(e.Pattern) {
		return
	}
	newPattern := append([]PatternRow{}, e.Pattern...)
	mutate(&newPattern[index])
	if e.OnPatternChange != nil {
		e.OnPatternChange(newPattern)
	}
}

func (e *ScheduleEditor) removePatternRow(index int) {
	if index < 0 || index >= len(e.Pattern) {
		return
	}
	newPattern := make([]PatternRow, 0, len(e.Pattern)-1)
	for i, r := range e.Pattern {
		if i != index {
			newPattern = append(newPattern, r)
		}
	}
	if e.OnPatternChange != nil {
		e.OnPatternChange(newPattern)
	}
}

func patternRowInvalid(index int, row PatternRow, all []PatternRow) bool {
	if len(row.Days) == 0 {
		return true
	}
	if row.StartMin >= row.EndMin {
		return true
	}
	for i, other := range all {
		if i == index {
			continue
		}
		if sharesDay(row.Days, other.Days) && rangesOverlap(row.StartMin, row.EndMin, other.StartMin, other.EndMin) {
			return true
		}
	}
	return false
}

func sharesDay(a, b []int) bool {
	for _, x := range a {
		if containsInt(b, x) {
			return true
		}
	}
	return false
}

func rangesOverlap(s1, e1, s2, e2 int) bool {
	return s1 < e2 && s2 < e1
}

func containsInt(slice []int, v int) bool {
	for _, x := range slice {
		if x == v {
			return true
		}
	}
	return false
}

func removeInt(slice []int, v int) []int {
	out := make([]int, 0, len(slice))
	for _, x := range slice {
		if x != v {
			out = append(out, x)
		}
	}
	return out
}

// ---------------------------------------------------------------------------
// Marcador de Días (Day Marker)
// ---------------------------------------------------------------------------

func (e *ScheduleEditor) horizonMonths() int {
	if e.Horizon <= 0 {
		return 6
	}
	return e.Horizon
}

func (e *ScheduleEditor) buildMarker() *Element {
	marker := Div().Set(clsMarker.AsAttr())
	marker.Child(Div().Text(lang.Translate("Marked days").String()))

	// Common hours controls
	hoursRow := Div().Set(clsMarkerHours.AsAttr())
	hoursRow.Child(Span().Text(lang.Translate("Hours for marked days").String()))

	startSel := NewElement("select").Attr("name", "marker-start").Bind(e.markerStart)
	for _, opt := range hourOptions(signalMinutes(e.markerStart), e.Bounds, 15) {
		startSel.Child(opt)
	}
	startSel.On("change", func(ev Event) {
		e.markerStart.Set(ev.TargetValue())
	})
	hoursRow.Child(startSel)

	endSel := NewElement("select").Attr("name", "marker-end").Bind(e.markerEnd)
	for _, opt := range hourOptions(signalMinutes(e.markerEnd), e.Bounds, 15) {
		endSel.Child(opt)
	}
	endSel.On("change", func(ev Event) {
		e.markerEnd.Set(ev.TargetValue())
	})
	hoursRow.Child(endSel)

	marker.Child(hoursRow)

	// Combine Holidays & Closures into read-only unavailable dates
	unavailHolidays := make([]string, 0, len(e.Holidays)+len(e.Closures))
	unavailHolidays = append(unavailHolidays, e.Holidays...)
	unavailHolidays = append(unavailHolidays, e.Closures...)

	// Occupation list for calendarslider: all dates in horizon or marked days get occupation=100 so they are selectable
	occ := occupationFromMarkedAndHorizon(e.Marked, e.horizonMonths())

	cal := &calendarslider.CalendarSlider{
		NumMonths:    e.horizonMonths(),
		Holidays:     toCalHolidays(unavailHolidays),
		Occupation:   occ,
		SelectedMany: e.markedSel,
		Expanded:     NewBool(false),
		OnToggle: func(date string, selected bool) {
			e.handleDayToggle(date, selected)
		},
	}

	marker.Child(cal)

	// List of marked days with individual hour pickers if divergent
	if len(e.Marked) > 0 {
		markedList := Ul().Set(clsExcList.AsAttr())
		for _, md := range e.Marked {
			mDay := md
			item := Li().Set(clsExcItem.AsAttr()).Key("marked-" + mDay.Date)
			item.Child(Span().Text(mDay.Date))

			mStartSel := NewElement("select").Attr("name", "md-start-"+mDay.Date)
			for _, opt := range hourOptions(mDay.StartMin, e.Bounds, 15) {
				mStartSel.Child(opt)
			}
			mStartSel.On("change", func(ev Event) {
				min, err := fmt.Convert(ev.TargetValue()).Int()
				if err == nil && e.OnMarkedDayEdit != nil {
					e.OnMarkedDayEdit(MarkedDay{Date: mDay.Date, StartMin: min, EndMin: mDay.EndMin})
				}
			})

			mEndSel := NewElement("select").Attr("name", "md-end-"+mDay.Date)
			for _, opt := range hourOptions(mDay.EndMin, e.Bounds, 15) {
				mEndSel.Child(opt)
			}
			mEndSel.On("change", func(ev Event) {
				min, err := fmt.Convert(ev.TargetValue()).Int()
				if err == nil && e.OnMarkedDayEdit != nil {
					e.OnMarkedDayEdit(MarkedDay{Date: mDay.Date, StartMin: mDay.StartMin, EndMin: min})
				}
			})

			item.Child(mStartSel).Child(mEndSel)
			markedList.Child(item)
		}
		marker.Child(markedList)
	}

	return marker
}

func occupationFromMarkedAndHorizon(marked []MarkedDay, horizonMonths int) []calendarslider.OccupationDay {
	startYear, startMonth := date.ParseMonthKey(time.FormatDate(time.Now())[:7])
	if startYear == 0 {
		startYear = 2026
		startMonth = 1
	}

	out := make([]calendarslider.OccupationDay, 0, horizonMonths*31)
	seen := make([]string, 0, horizonMonths*31)

	for m := 0; m < horizonMonths; m++ {
		y, mon := date.AddMonths(startYear, startMonth, m)
		days := date.DaysInMonth(y, mon)
		for d := 1; d <= days; d++ {
			dStr := date.DateKey(y, mon, d)
			out = append(out, calendarslider.OccupationDay{Date: dStr, Percent: 50})
			seen = append(seen, dStr)
		}
	}

	for _, md := range marked {
		if !containsDate(seen, md.Date) {
			out = append(out, calendarslider.OccupationDay{Date: md.Date, Percent: 100})
		}
	}
	return out
}

func (e *ScheduleEditor) handleDayToggle(dateStr string, selected bool) {
	if selected {
		startMin := signalMinutes(e.markerStart)
		endMin := signalMinutes(e.markerEnd)
		if e.OnDaysMarked != nil {
			e.OnDaysMarked([]string{dateStr}, startMin, endMin)
		}
	} else {
		if e.OnDaysUnmarked != nil {
			e.OnDaysUnmarked([]string{dateStr})
		}
	}
}

// ---------------------------------------------------------------------------
// Panel de excepciones
// ---------------------------------------------------------------------------

func (e *ScheduleEditor) buildExceptions() *Element {
	return Div().Set(clsExceptions.AsAttr()).
		Child(&calendarslider.CalendarSlider{
			NumMonths:  3,
			Holidays:   toCalHolidays(e.Holidays),
			Occupation: occupationFromExceptions(e.Exceptions),
			Selected:   e.sel,
			OnSelect:   func(dateKey string) { e.sel.Set(dateKey) },
		}).
		Child(e.buildExceptionForm()).
		Child(e.buildExceptionList())
}

func toCalHolidays(days []string) []calendarslider.Holiday {
	out := make([]calendarslider.Holiday, 0, len(days))
	for _, d := range days {
		out = append(out, calendarslider.Holiday{Date: d, Name: "Feriado"})
	}
	return out
}

func occupationFromExceptions(excs []Exception) []calendarslider.OccupationDay {
	out := make([]calendarslider.OccupationDay, 0, len(excs))
	for _, ex := range excs {
		p := 50
		if ex.Type == ExcHoliday || ex.Type == ExcBlocked {
			p = 100
		}
		out = append(out, calendarslider.OccupationDay{Date: ex.Date, Percent: p})
	}
	return out
}

func (e *ScheduleEditor) buildExceptionForm() *Element {
	form := Div().Set(clsExcForm.AsAttr()).
		BindStateFunc(widget.Open, func() bool { return e.sel.Get() != "" })

	typeRow := Div().Set(clsExcType.AsAttr())
	typeRow.Child(Span().Text(lang.Translate("Type").String()))
	for _, opt := range exceptionTypes() {
		radio := Input("radio").Set(clsExcType.AsAttr()).
			Key(opt.Key).
			Attr("name", "scheduleeditor-type").
			Attr("value", opt.Key).
			BindAttrBoolFunc("checked", func() bool { return e.excType.Get() == opt.Key })
		radio.On("change", func(ev Event) {
			e.excType.Set(ev.TargetValue())
			if opt.Key == ExcHoliday {
				e.excFrom.Set("540")
				e.excTo.Set("1080")
			}
		})
		label := Label().For(radio).Text(opt.Value)
		typeRow.Child(radio).Child(label)
	}

	hoursRow := Div().Set(clsExcHours.AsAttr()).
		BindStateFunc(widget.Open, func() bool { return e.excType.Get() != ExcHoliday }).
		Child(boundTimePick(PartExcHours, "exc-from", e.excFrom, e.Bounds)).
		Child(boundTimePick(PartExcHours, "exc-to", e.excTo, e.Bounds))

	notes := Input("text").Set(clsExcNotes.AsAttr()).
		Attr("name", "scheduleeditor-notes").
		Attr("placeholder", lang.Translate("Notes").String()).
		Bind(e.excNote)

	add := Button().Set(clsExcAdd.AsAttr()).
		Attr("type", "button").
		Text(lang.Translate("Add").String())
	add.On("click", func(Event) { e.addException() })

	dateRow := Div().Child(Span().Text(lang.Translate("Date").String())).
		Child(Span().BindText(e.sel))

	return form.
		Child(dateRow).
		Child(typeRow).
		Child(hoursRow).
		Child(notes).
		Child(add)
}

func (e *ScheduleEditor) addException() {
	if e.sel.Get() == "" {
		return
	}
	exc := Exception{
		ID:    "",
		Date:  e.sel.Get(),
		Type:  e.excType.Get(),
		Notes: e.excNote.Get(),
	}
	if exc.Type == ExcSpecialHours || exc.Type == ExcBlocked {
		exc.StartMin = signalMinutes(e.excFrom)
		exc.EndMin = signalMinutes(e.excTo)
	}
	if e.OnExceptionAdd != nil {
		e.OnExceptionAdd(exc)
	}
	e.sel.Set("")
	e.excNote.Set("")
}

func exceptionTypes() []fmt.KeyValue {
	return []fmt.KeyValue{
		{Key: ExcHoliday, Value: lang.Translate("Closed").String()},
		{Key: ExcSpecialHours, Value: lang.Translate("Special hours").String()},
		{Key: ExcBlocked, Value: lang.Translate("Blocked").String()},
	}
}

func signalMinutes(s *SignalString) int {
	if s == nil {
		return 0
	}
	m, err := fmt.Convert(s.Get()).Int()
	if err != nil {
		return 0
	}
	return m
}

func boundTimePick(part widget.Part, name string, sig *SignalString, b Bounds) *Element {
	sel := NewElement("select").
		Set(NameScheduleEditor.Class(part).AsAttr()).
		Attr("name", name).
		Bind(sig)
	for _, opt := range hourOptions(signalMinutes(sig), b, 15) {
		sel.Child(opt)
	}
	sel.On("change", func(ev Event) {
		sig.Set(ev.TargetValue())
	})
	return sel
}

func (e *ScheduleEditor) buildExceptionList() *Element {
	list := Ul().Set(clsExcList.AsAttr())
	items := sortedExceptions(e.Exceptions)

	if len(items) == 0 && len(e.Holidays) == 0 {
		list.Child(Li().Set(clsExcItem.AsAttr()).
			Text(lang.Translate("No exceptions").String()))
		return list
	}

	for _, ex := range items {
		isHoliday := containsDate(e.Holidays, ex.Date)
		row := Li().Set(clsExcItem.AsAttr()).Key(ex.Date + "/" + ex.Type)

		row.Child(Span().Text(ex.Date)).
			Child(Span().Text(exceptionLabel(ex.Type)))
		if hours, ok := exceptionHoursText(ex); ok {
			row.Child(Span().Text(hours))
		}
		if ex.Notes != "" {
			row.Child(Span().Set(clsExcNotes.AsAttr()).Text(ex.Notes))
		}

		if isHoliday {
			row.Set(clsExcHoliday.AsAttr())
		} else {
			remove := Button().Set(clsExcRemove.AsAttr()).
				Attr("type", "button").
				Text(lang.Translate("Remove").String())
			id := ex.ID
			remove.On("click", func(Event) {
				if e.OnExceptionRemove != nil {
					e.OnExceptionRemove(id)
				}
			})
			row.Child(remove)
		}
		list.Child(row)
	}
	return list
}

func sortedExceptions(excs []Exception) []Exception {
	out := append([]Exception{}, excs...)
	for i := 0; i < len(out); i++ {
		min := i
		for j := i + 1; j < len(out); j++ {
			if out[j].Date < out[min].Date {
				min = j
			}
		}
		if min != i {
			out[i], out[min] = out[min], out[i]
		}
	}
	return out
}

func exceptionLabel(t string) string {
	switch t {
	case ExcHoliday:
		return lang.Translate("Closed").String()
	case ExcSpecialHours:
		return lang.Translate("Special hours").String()
	case ExcBlocked:
		return lang.Translate("Blocked").String()
	default:
		return t
	}
}

func exceptionHoursText(ex Exception) (string, bool) {
	if ex.Type != ExcSpecialHours && ex.Type != ExcBlocked {
		return "", false
	}
	return hhmm(ex.StartMin) + "–" + hhmm(ex.EndMin), true
}

func containsDate(list []string, date string) bool {
	for _, d := range list {
		if d == date {
			return true
		}
	}
	return false
}

func (e *ScheduleEditor) Render() *Element {
	return Div().Set(clsRoot.AsAttr()).
		Child(e.buildPattern()).
		Child(e.buildMarker()).
		Child(e.buildExceptions())
}
