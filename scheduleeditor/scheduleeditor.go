// Package scheduleeditor is a pure component for editing a professional's
// weekly schedule template plus per-date exceptions. It knows nothing about
// router, orm or appointment_booking: the host supplies the data (Week,
// Exceptions, Holidays) and translates the callbacks to persistence ops.
//
// Time shape = minutes from midnight (int), the same shape the
// appointment_booking module persists (work_calendar_weekly). Times are
// presented in HH:MM inside select options; the value carried is the integer.
package scheduleeditor

import (
	"webtyp.com/date"
	. "webtyp.com/dom"
	"webtyp.com/fmt"
	"webtyp.com/fmt/lang"
	. "webtyp.com/html"
	"webtyp.com/widget"

	"webtyp.com/components/calendarslider"
)

// NameScheduleEditor is the widget identity.
const NameScheduleEditor = widget.Name("scheduleeditor")

const (
	PartWeek       = widget.Part("week")
	PartWeekRow    = widget.Part("week-row")
	PartDay        = widget.Part("day")
	PartDayName    = widget.Part("day-name")
	PartToggle     = widget.Part("toggle")
	PartTime       = widget.Part("time")
	PartExceptions = widget.Part("exceptions")
	PartSlider     = widget.Part("slider")
	PartExcForm    = widget.Part("exc-form")
	PartExcHours   = widget.Part("exc-hours")
	PartExcType    = widget.Part("exc-type")
	PartExcNotes   = widget.Part("exc-notes")
	PartExcAdd     = widget.Part("exc-add")
	PartExcList    = widget.Part("exc-list")
	PartExcItem    = widget.Part("exc-item")
	PartExcHoliday = widget.Part("exc-holiday")
	PartExcRemove  = widget.Part("exc-remove")
)

var (
	clsRoot       = NameScheduleEditor.Root()
	clsWeek       = NameScheduleEditor.Class(PartWeek)
	clsWeekRow    = NameScheduleEditor.Class(PartWeekRow)
	clsDay        = NameScheduleEditor.Class(PartDay)
	clsDayName    = NameScheduleEditor.Class(PartDayName)
	clsToggle     = NameScheduleEditor.Class(PartToggle)
	clsTime       = NameScheduleEditor.Class(PartTime)
	clsExceptions = NameScheduleEditor.Class(PartExceptions)
	clsSlider     = NameScheduleEditor.Class(PartSlider)
	clsExcForm    = NameScheduleEditor.Class(PartExcForm)
	clsExcHours   = NameScheduleEditor.Class(PartExcHours)
	clsExcType    = NameScheduleEditor.Class(PartExcType)
	clsExcNotes   = NameScheduleEditor.Class(PartExcNotes)
	clsExcAdd     = NameScheduleEditor.Class(PartExcAdd)
	clsExcList    = NameScheduleEditor.Class(PartExcList)
	clsExcItem    = NameScheduleEditor.Class(PartExcItem)
	clsExcHoliday = NameScheduleEditor.Class(PartExcHoliday)
	clsExcRemove  = NameScheduleEditor.Class(PartExcRemove)
)

// Tipos de excepción (valor de Exception.Type). Son las claves que
// appointment_booking espera en work_calendar_exception.exception_type.
const (
	ExcHoliday      = "HOLIDAY"
	ExcSpecialHours = "SPECIAL_HOURS"
	ExcBlocked      = "BLOCKED"
)

// WeeklyRow es una fila de la plantilla semanal. DayOfWeek: 0=Domingo …
// 6=Sábado. Los tiempos son minutos desde medianoche (0..1439); colación
// 0/0 = sin colación.
type WeeklyRow struct {
	DayOfWeek               int
	Active                  bool
	WorkStart, WorkFinish   int
	BreakStart, BreakFinish int
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

// ScheduleEditor edita la agenda de un profesional: plantilla semanal de 7
// filas + panel de excepciones. Componente puro: no conoce router/orm/… — el
// host traduce los callbacks a las ops de appointment_booking.
type ScheduleEditor struct {
	Element // value embed — NEVER pointer

	// Week: las 7 filas (Dom..Sáb) ya ordenadas. Estado INICIAL — el host
	// persiste y remonta con datos frescos; el componente no es la fuente de
	// verdad.
	Week []WeeklyRow
	// Exceptions: excepciones vigentes. Estado inicial, misma regla.
	Exceptions []Exception
	// Holidays: fechas de feriado nacional "YYYY-MM-DD", solo lectura.
	Holidays []string

	OnWeeklyChange    func(WeeklyRow) // fila editada (toggle/entrada/salida/colación)
	OnExceptionAdd    func(Exception) // alta desde el panel (ID == "")
	OnExceptionRemove func(id string)

	sel     *SignalString // día elegido por el calendario ("" = form oculto)
	excType *SignalString // tipo elegido en el formulario de alta
	excFrom *SignalString // hora "desde" del formulario (minutos)
	excTo   *SignalString // hora "hasta" del formulario (minutos)
	excNote *SignalString // notas del formulario
}

func (e *ScheduleEditor) WidgetName() widget.Name { return NameScheduleEditor }
func (e *ScheduleEditor) WidgetKind() widget.Kind { return widget.Form }

func (e *ScheduleEditor) Init(_ Ctx) {
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

// hourOptions devuelve las opciones de un <select> de hora: 06:00 a 22:00
// en pasos de 15 minutos. value = minutos (string), texto = HH:MM.
func hourOptions(selected int) []*Element {
	opts := make([]*Element, 0, 65)
	for m := 360; m <= 1320; m += 15 {
		if m == selected {
			opts = append(opts, SelectedOption(fmt.Convert(m).String(), hhmm(m)))
		} else {
			opts = append(opts, Option(fmt.Convert(m).String(), hhmm(m)))
		}
	}
	return opts
}

// ---------------------------------------------------------------------------
// Plantilla semanal
// ---------------------------------------------------------------------------

func (e *ScheduleEditor) buildWeek() *Element {
	week := Div().Set(clsWeek.AsAttr()).Attr("role", "grid")
	for i, r := range e.Week {
		week.Child(e.buildWeekRow(i, r))
	}
	return week
}

func (e *ScheduleEditor) buildWeekRow(i int, r WeeklyRow) *Element {
	invalid := weekInvalid(r)
	if invalid {
		Log("scheduleeditor: weekly row for", date.WeekdayName(r.DayOfWeek), "invalid (work or break window)")
	}

	active := NewBool(r.Active)
	row := Div().Set(clsWeekRow.AsAttr()).
		Attr("data-active", mapBool(r.Active)).
		BindState(widget.Invalid, NewBool(invalid))

	// Toggle: set Active y dispara OnWeeklyChange con la fila resultante.
	toggle := Input("checkbox").Set(clsToggle.AsAttr()).
		BindAttrBool("checked", active)
	toggle.On("change", func(ev Event) {
		e.weeklyChange(i, func(r *WeeklyRow) { r.Active = ev.TargetChecked() })
	})

	window := Div().Set(clsDay.AsAttr()).
		Child(Span().Set(clsDayName.AsAttr()).Text(dayLabel(r.DayOfWeek))).
		Child(toggle)
	row.Child(window)

	row.Child(timePick(PartTime, "work-start", r.WorkStart, func(m int) {
		e.weeklyChange(i, func(r *WeeklyRow) { r.WorkStart = m })
	}))
	row.Child(timePick(PartTime, "work-finish", r.WorkFinish, func(m int) {
		e.weeklyChange(i, func(r *WeeklyRow) { r.WorkFinish = m })
	}))
	row.Child(timePick(PartTime, "break-start", r.BreakStart, func(m int) {
		e.weeklyChange(i, func(r *WeeklyRow) { r.BreakStart = m })
	}))
	row.Child(timePick(PartTime, "break-finish", r.BreakFinish, func(m int) {
		e.weeklyChange(i, func(r *WeeklyRow) { r.BreakFinish = m })
	}))

	return row
}

// weeklyChange es el embudo único de cada edición de la plantilla semanal:
// recompone la fila desde Week[i] (la fila editada), aplica la mutación y
// dispara OnWeeklyChange una sola vez. Método para que el DOM y los tests
// compartan exactamente el mismo camino.
func (e *ScheduleEditor) weeklyChange(i int, mutate func(*WeeklyRow)) {
	if i < 0 || i >= len(e.Week) {
		Log("scheduleeditor: weekly row index out of range", i)
		return
	}
	r := e.Week[i]
	mutate(&r)
	if e.OnWeeklyChange != nil {
		e.OnWeeklyChange(r)
	}
}

// dayLabel es el nombre del día. El componente es librería → renderiza el
// nombre canónico inglés de webtyp/date (Sunday..Saturday) vía lang.Translate:
// la app registra el diccionario, el componente solo pide traducir.
func dayLabel(dow int) string {
	return lang.Translate(date.WeekdayName(dow)).String()
}

// timePick arma un <select> de hora con las opciones 06:00–22:00, su valor
// inicial y un listener de cambio que entrega los minutos al callback.
func timePick(part widget.Part, name string, val int, onChange func(int)) *Element {
	sel := NewElement("select").
		Set(NameScheduleEditor.Class(part).AsAttr()).
		Attr("name", name)
	for _, opt := range hourOptions(val) {
		sel.Child(opt)
	}
	sel.On("change", func(ev Event) {
		m, err := fmt.Convert(ev.TargetValue()).Int()
		if err != nil || m < 0 || m > 1439 {
			Log("scheduleeditor: bad time value", ev.TargetValue())
			return
		}
		onChange(m)
	})
	return sel
}

// weekInvalid marca una fila inválida (dev-warning no bloqueante): WorkStart <
// WorkFinish; con colación, WorkStart <= BreakStart < BreakFinish <=
// WorkFinish. Una fila con 0/0 (sin ventana aún, p. ej. un día inactivo) no es
// inválida: es un día no configurado.
func weekInvalid(r WeeklyRow) bool {
	if r.WorkStart == 0 && r.WorkFinish == 0 {
		return false
	}
	if r.WorkStart >= r.WorkFinish {
		return true
	}
	hasBreak := r.BreakStart != 0 || r.BreakFinish != 0
	if !hasBreak {
		return false
	}
	return r.BreakStart < r.WorkStart || r.BreakStart >= r.BreakFinish || r.BreakFinish > r.WorkFinish
}

func mapBool(b bool) string {
	if b {
		return "true"
	}
	return "false"
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

// toCalHolidays mapea el []string "YYYY-MM-DD" de feriados del componente al
// type que calendarslider espera (todo con la misma etiqueta "Feriado").
func toCalHolidays(days []string) []calendarslider.Holiday {
	out := make([]calendarslider.Holiday, 0, len(days))
	for _, d := range days {
		out = append(out, calendarslider.Holiday{Date: d, Name: "Feriado"})
	}
	return out
}

// occupationFromExceptions marca los días con excepción como seleccionables en
// el calendario (su regla: día con Occupation = clicable) y los pinta con el
// uso: 100 para HOLIDAY/BLOCKED, 50 para SPECIAL_HOURS. Solo es un marcador —
// la realidad vive en la lista de excepciones.
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

// buildExceptionForm es el formulario de alta inline, visible solo cuando hay
// un día elegido (e.sel != ""). La visibilidad se ata al signal, sin
// reconstruir el árbol.
func (e *ScheduleEditor) buildExceptionForm() *Element {
	form := Div().Set(clsExcForm.AsAttr()).
		BindAttrBoolFunc("data-open", func() bool { return e.sel.Get() != "" })

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
				// HOLIDAY no usa ventana de horas; reset a un valor sano.
				e.excFrom.Set("540")
				e.excTo.Set("1080")
			}
		})
		label := Label().For(radio).Text(opt.Value)
		typeRow.Child(radio).Child(label)
	}

	// Los selects de hora solo aplican a SPECIAL_HOURS/BLOCKED — se ocultan
	// para HOLIDAY (bind sobre e.excType, sin reconstruir).
	hoursRow := Div().Set(clsExcHours.AsAttr()).
		BindAttrBoolFunc("data-open", func() bool { return e.excType.Get() != ExcHoliday }).
		Child(boundTimePick(PartTime, "exc-from", e.excFrom)).
		Child(boundTimePick(PartTime, "exc-to", e.excTo))

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

// addException es la acción del botón "Agregar": arma la Exception desde los
// signals del formulario, invoca OnExceptionAdd y limpia el día elegido. Es un
// método para que un test pueda dispararla sin simular el clic en el DOM.
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
		{Key: ExcSpecialHours, Value: lang.Translate("Special", "hours").String()},
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

// boundTimePick es un <select> de hora cuyo valor vive en un SignalString
// (dos vías), para el formulario de alta de excepciones.
func boundTimePick(part widget.Part, name string, sig *SignalString) *Element {
	sel := NewElement("select").
		Set(NameScheduleEditor.Class(part).AsAttr()).
		Attr("name", name).
		Bind(sig)
	for _, opt := range hourOptions(signalMinutes(sig)) {
		sel.Child(opt)
	}
	sel.On("change", func(ev Event) {
		sig.Set(ev.TargetValue())
	})
	return sel
}

// buildExceptionList renderiza las excepciones vigentes en orden de fecha
// ascendente. Las de Holidays son solo-lectura (sin "Quitar").
func (e *ScheduleEditor) buildExceptionList() *Element {
	list := Ul().Set(clsExcList.AsAttr())
	items := sortedExceptions(e.Exceptions)

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

// sortedExceptions ordena por fecha ascendente sin `sort` de la stdlib —
// selección lineal decidida: el conjunto de excepciones es pequeño.
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
		return lang.Translate("Special", "hours").String()
	case ExcBlocked:
		return lang.Translate("Blocked").String()
	default:
		return t
	}
}

// exceptionHoursText devuelve el texto "HH:MM–HH:MM" para SPECIAL_HOURS/
// BLOCKED (false para el resto).
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
	if len(e.Week) != 7 {
		Log("scheduleeditor: expected 7 weekly rows (Sun..Sat), got", len(e.Week))
	}
	return Div().Set(clsRoot.AsAttr()).
		Child(e.buildWeek()).
		Child(e.buildExceptions())
}
