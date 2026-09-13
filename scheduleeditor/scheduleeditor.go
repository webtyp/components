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
	PartPattern      = widget.Part("pattern")
	PartPatternRow   = widget.Part("pattern-row")
	PartDayChips     = widget.Part("day-chips")
	PartDayChip      = widget.Part("day-chip")
	PartDayLabel     = widget.Part("day-label")
	PartFieldLabel   = widget.Part("field-label")
	PartRowRemove    = widget.Part("row-remove")
	PartRowAdd       = widget.Part("row-add")
	PartMarker       = widget.Part("marker")
	PartMarkerHours  = widget.Part("marker-hours")
	PartSectionTitle = widget.Part("section-title")
	PartTimeRange    = widget.Part("time-range")
	PartDayRow       = widget.Part("day-row")
	PartDaySlots     = widget.Part("day-slots")
	PartDayOffText   = widget.Part("day-off-text")
	PartSlotAdd      = widget.Part("slot-add")
	PartSlotRemove   = widget.Part("slot-remove")
	PartHintText     = widget.Part("hint-text")
	PartExceptions   = widget.Part("exceptions")
	PartSlider       = widget.Part("slider")
	PartExcForm      = widget.Part("exc-form")
	PartExcHours     = widget.Part("exc-hours")
	PartExcNotes     = widget.Part("exc-notes")
	PartExcAdd       = widget.Part("exc-add")
	PartExcList      = widget.Part("exc-list")
	PartExcItem      = widget.Part("exc-item")
	PartExcDate      = widget.Part("exc-date")
	PartExcDateRow   = widget.Part("exc-date-row")
	PartExcAction    = widget.Part("exc-action")
	PartExcHint      = widget.Part("exc-hint")
	PartExcHoliday   = widget.Part("exc-holiday")
	PartExcRemove    = widget.Part("exc-remove")
)

var (
	clsRoot         = NameScheduleEditor.Root()
	clsPattern      = NameScheduleEditor.Class(PartPattern)
	clsPatternRow   = NameScheduleEditor.Class(PartPatternRow)
	clsDayChips     = NameScheduleEditor.Class(PartDayChips)
	clsDayChip      = NameScheduleEditor.Class(PartDayChip)
	clsDayLabel     = NameScheduleEditor.Class(PartDayLabel)
	clsFieldLabel   = NameScheduleEditor.Class(PartFieldLabel)
	clsRowRemove    = NameScheduleEditor.Class(PartRowRemove)
	clsRowAdd       = NameScheduleEditor.Class(PartRowAdd)
	clsMarker       = NameScheduleEditor.Class(PartMarker)
	clsMarkerHours  = NameScheduleEditor.Class(PartMarkerHours)
	clsSectionTitle = NameScheduleEditor.Class(PartSectionTitle)
	clsTimeRange    = NameScheduleEditor.Class(PartTimeRange)
	clsDayRow       = NameScheduleEditor.Class(PartDayRow)
	clsDaySlots     = NameScheduleEditor.Class(PartDaySlots)
	clsDayOffText   = NameScheduleEditor.Class(PartDayOffText)
	clsSlotAdd      = NameScheduleEditor.Class(PartSlotAdd)
	clsSlotRemove   = NameScheduleEditor.Class(PartSlotRemove)
	clsHintText     = NameScheduleEditor.Class(PartHintText)
	clsExceptions   = NameScheduleEditor.Class(PartExceptions)
	clsSlider       = NameScheduleEditor.Class(PartSlider)
	clsExcForm      = NameScheduleEditor.Class(PartExcForm)
	clsExcHours     = NameScheduleEditor.Class(PartExcHours)
	clsExcNotes     = NameScheduleEditor.Class(PartExcNotes)
	clsExcAdd       = NameScheduleEditor.Class(PartExcAdd)
	clsExcList      = NameScheduleEditor.Class(PartExcList)
	clsExcItem      = NameScheduleEditor.Class(PartExcItem)
	clsExcDate      = NameScheduleEditor.Class(PartExcDate)
	clsExcDateRow   = NameScheduleEditor.Class(PartExcDateRow)
	clsExcAction    = NameScheduleEditor.Class(PartExcAction)
	clsExcHint      = NameScheduleEditor.Class(PartExcHint)
	clsExcHoliday   = NameScheduleEditor.Class(PartExcHoliday)
	clsExcRemove    = NameScheduleEditor.Class(PartExcRemove)
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

	sel     *SignalString // día elegido por el calendario ("" = form oculto)
	calOpen *SignalBool   // el calendario queda desplegado: acá es la pantalla, no un filtro
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
	if e.calOpen == nil {
		e.calOpen = NewBool(true)
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

// dayFullLabel es el nombre largo del día. La fila por día lo lleva entero
// ("Lunes"): una lista de siete renglones tiene sitio para la palabra, y la
// abreviatura solo existía porque los chips competían por el ancho de una fila.
func dayFullLabel(index int) string {
	if index < 0 || index > 6 {
		return ""
	}
	return lang.Translate(date.WeekdayName(index)).String()
}

// ---------------------------------------------------------------------------
// Patrón semanal — una fila POR DÍA
// ---------------------------------------------------------------------------
//
// La sección lista los siete días SIEMPRE, cada uno con su interruptor y sus
// rangos. Antes listaba "filas" abstractas: un rango horario más un juego de
// casillas eligiendo a qué días aplicaba.
//
// Por qué cambió. La fila abstracta contesta "¿a qué días aplica este bloque
// de 08:00–14:00?", y nadie planifica así: la pregunta real es "¿qué horario
// tengo los lunes?". De ahí salían tres defectos que ninguna cantidad de CSS
// arregla:
//
//  1. "Quitar fila" no significaba nada — no hay tal fila en la cabeza de
//     quien lo usa, es un artefacto de cómo se guarda el dato.
//  2. Un día libre era la AUSENCIA de una fila que lo mencionara: había que
//     cruzar todas las filas mentalmente para saber si el martes se trabaja.
//     Con los siete días a la vista, un día libre se ve.
//  3. El solapamiento podía ocurrir ENTRE filas — el mismo día en dos reglas
//     distintas — y detectarlo obligaba a cruzar el patrón entero. Por día
//     solo puede ocurrir entre los rangos de ESE día: una comprobación local
//     y explicable, "estos dos horarios del lunes se pisan".
//
// Es el patrón de Calendly, Cal.com, Acuity, Google Calendar (working hours),
// Microsoft Bookings, Square Appointments y Fresha: "weekly hours" es una
// lista de los siete días, no una lista de reglas.
//
// El contrato con el host NO cambia: se sigue emitiendo []PatternRow por
// OnPatternChange. Lo único es que al primer cambio las filas se normalizan a
// una por (día, rango) — semánticamente idéntico, y es lo que hace que el
// mapeo día↔dato sea trivial y sin ambigüedad.

// dayRange es un rango de un día concreto, ya desanidado del PatternRow.
type dayRange struct {
	StartMin, EndMin int
}

// rangesForDay reúne los rangos que hoy cubren el día d, en orden de entrada.
func (e *ScheduleEditor) rangesForDay(d int) []dayRange {
	out := make([]dayRange, 0, 2)
	for _, row := range e.Pattern {
		if containsInt(row.Days, d) {
			out = append(out, dayRange{StartMin: row.StartMin, EndMin: row.EndMin})
		}
	}
	return out
}

// emitDayRanges reconstruye el patrón con los rangos de d reemplazados por
// next (vacío = el día no se trabaja) y lo emite.
func (e *ScheduleEditor) emitDayRanges(d int, next []dayRange) {
	rows := make([]PatternRow, 0, len(e.Pattern)*2)
	for _, wd := range date.WeekOrder() {
		src := e.rangesForDay(wd)
		if wd == d {
			src = next
		}
		for _, r := range src {
			rows = append(rows, PatternRow{StartMin: r.StartMin, EndMin: r.EndMin, Days: []int{wd}})
		}
	}
	e.Pattern = rows
	if e.OnPatternChange != nil {
		e.OnPatternChange(rows)
	}
}

func (e *ScheduleEditor) buildPattern() *Element {
	container := Div().Set(clsPattern.AsAttr())
	container.Child(Div().Set(clsSectionTitle.AsAttr()).Text(lang.Translate("Weekly pattern").String()))
	container.Child(Div().Set(clsHintText.AsAttr()).
		Text(lang.Translate("Turn on the days you work and set their hours").String()))

	for _, d := range date.WeekOrder() {
		container.Child(e.buildDayRow(d))
	}
	return container
}

// buildDayRow es la fila de UN día de la semana: nombre, interruptor, y sus
// rangos. Siempre presente, se trabaje ese día o no.
func (e *ScheduleEditor) buildDayRow(d int) *Element {
	dVal := d
	ranges := e.rangesForDay(dVal)
	on := len(ranges) > 0

	rowEl := Div().Set(clsDayRow.AsAttr()).
		Key("day-"+fmt.Convert(dVal).String()).
		BindState(widget.Selected, NewBool(on)).
		BindState(widget.Invalid, NewBool(dayInvalid(ranges)))

	toggle := Input("checkbox").Set(clsDayChip.AsAttr()).
		Key("toggle-"+fmt.Convert(dVal).String()).
		BindAttrBool("checked", NewBool(on))
	toggle.OnChange(func(ev Event) {
		if ev.TargetChecked() {
			e.emitDayRanges(dVal, []dayRange{e.defaultRange()})
		} else {
			e.emitDayRanges(dVal, nil)
		}
	})

	// El <label> es la píldora visible; el checkbox real queda accesible pero
	// fuera de vista (un checkbox nativo no se puede pintar).
	name := Label().For(toggle).
		Set(clsDayLabel.AsAttr()).
		BindState(widget.Selected, NewBool(on)).
		Text(dayFullLabel(dVal))

	rowEl.Child(toggle).Child(name)

	if !on {
		// Un día libre lo dice con todas las letras. Antes había que deducirlo
		// de que ninguna fila lo mencionara.
		rowEl.Child(Span().Set(clsDayOffText.AsAttr()).
			Text(lang.Translate("Does not work").String()))
		return rowEl
	}

	slots := Div().Set(clsDaySlots.AsAttr())
	for i, r := range ranges {
		slots.Child(e.buildDaySlot(dVal, i, r, len(ranges) > 1))
	}

	// Un segundo rango en el mismo día es la jornada partida: mañana y tarde.
	// Solo se ofrece si hay hueco después del último — un "+" que agrega algo
	// que el servidor va a rechazar es peor que un "+" ausente.
	if next, ok := e.nextPeriod(ranges); ok {
		addSlot := Button().Set(clsSlotAdd.AsAttr()).
			Attr("type", "button").
			Attr("title", lang.Translate("Add time range").String()).
			Text("+")
		addSlot.OnClick(func(Event) {
			e.emitDayRanges(dVal, append(append([]dayRange{}, ranges...), next))
		})
		slots.Child(addSlot)
	}

	rowEl.Child(slots)
	return rowEl
}

// buildDaySlot es un rango horario dentro de un día: desde, hasta, y la
// papelera solo cuando hay más de uno (con uno solo, apagar el día es la
// acción correcta y la papelera sería un segundo camino para lo mismo).
func (e *ScheduleEditor) buildDaySlot(d, idx int, r dayRange, removable bool) *Element {
	dVal, iVal := d, idx
	slot := Div().Set(clsTimeRange.AsAttr()).
		Key("slot-" + fmt.Convert(dVal).String() + "-" + fmt.Convert(iVal).String())

	slot.Child(Span().Set(clsFieldLabel.AsAttr()).Text(lang.Translate("From").String()))
	startSel := NewElement("select").Attr("name", "start-time")
	for _, opt := range hourOptions(r.StartMin, e.Bounds, 15) {
		startSel.Child(opt)
	}
	startSel.OnChange(func(ev Event) {
		if m, err := fmt.Convert(ev.TargetValue()).Int(); err == nil {
			e.mutateSlot(dVal, iVal, func(x *dayRange) { x.StartMin = m })
		}
	})
	slot.Child(startSel)

	slot.Child(Span().Set(clsFieldLabel.AsAttr()).Text(lang.Translate("To").String()))
	endSel := NewElement("select").Attr("name", "end-time")
	for _, opt := range hourOptions(r.EndMin, e.Bounds, 15) {
		endSel.Child(opt)
	}
	endSel.OnChange(func(ev Event) {
		if m, err := fmt.Convert(ev.TargetValue()).Int(); err == nil {
			e.mutateSlot(dVal, iVal, func(x *dayRange) { x.EndMin = m })
		}
	})
	slot.Child(endSel)

	if removable {
		rm := Button().Set(clsSlotRemove.AsAttr()).
			Attr("type", "button").
			Attr("title", lang.Translate("Remove time range").String()).
			Text("×")
		rm.OnClick(func(Event) {
			cur := e.rangesForDay(dVal)
			if iVal < 0 || iVal >= len(cur) {
				return
			}
			e.emitDayRanges(dVal, append(append([]dayRange{}, cur[:iVal]...), cur[iVal+1:]...))
		})
		slot.Child(rm)
	}
	return slot
}

func (e *ScheduleEditor) mutateSlot(d, idx int, mutate func(*dayRange)) {
	cur := e.rangesForDay(d)
	if idx < 0 || idx >= len(cur) {
		return
	}
	mutate(&cur[idx])
	e.emitDayRanges(d, cur)
}

// dayInvalid informa si los rangos de UN día no se sostienen: alguno termina
// antes de empezar, o dos se pisan. Local a un día — no hay que cruzar el
// patrón entero, que es lo que hacía falta cuando una regla abarcaba varios
// días.
//
// El rango invertido va acá y no solo el solapamiento: la primera versión de
// esta comprobación solo miraba pares y dejaba pasar un 18:00–09:00 suelto,
// que la validación anterior sí atrapaba.
func dayInvalid(rs []dayRange) bool {
	for i, r := range rs {
		if r.EndMin <= r.StartMin {
			return true
		}
		for j := i + 1; j < len(rs); j++ {
			if r.StartMin < rs[j].EndMin && rs[j].StartMin < r.EndMin {
				return true
			}
		}
	}
	return false
}

// nextPeriod calcula el rango que agrega el "+": la tarde después de la
// mañana. Empieza una hora DESPUÉS de que termina el último — el corte de
// almuerzo — y dura lo mismo que él, recortado al cierre.
//
// No es defaultRange(). Ese es el horario de un día que se enciende desde
// cero; usarlo también acá agregaba un 09:00–18:00 encima de un 08:00–14:00
// que ya existía, el servidor lo rechazaba por solapamiento y el "+" no hacía
// nada visible. Un control que no puede producir un estado válido no debe
// ofrecerse: si no hay hueco, ok es false y el "+" no se dibuja.
func (e *ScheduleEditor) nextPeriod(ranges []dayRange) (dayRange, bool) {
	if len(ranges) == 0 {
		return e.defaultRange(), true
	}
	last := ranges[0]
	for _, r := range ranges {
		if r.EndMin > last.EndMin {
			last = r
		}
	}
	closeMin := 1440
	if e.Bounds.CloseMin > e.Bounds.OpenMin {
		closeMin = e.Bounds.CloseMin
	}

	const lunchBreak = 60
	start := last.EndMin + lunchBreak
	if start >= closeMin {
		return dayRange{}, false
	}
	dur := last.EndMin - last.StartMin
	end := start + dur
	if end > closeMin {
		end = closeMin
	}
	if end <= start {
		return dayRange{}, false
	}
	return dayRange{StartMin: start, EndMin: end}, true
}

// defaultRange es el horario con el que se enciende un día: la apertura del
// establecimiento si el host la declaró, y 09:00–18:00 si no.
func (e *ScheduleEditor) defaultRange() dayRange {
	if e.Bounds.CloseMin > e.Bounds.OpenMin {
		return dayRange{StartMin: e.Bounds.OpenMin, EndMin: e.Bounds.CloseMin}
	}
	return dayRange{StartMin: 540, EndMin: 1080}
}

func (e *ScheduleEditor) horizonMonths() int {
	if e.Horizon <= 0 {
		return 6
	}
	return e.Horizon
}

// ---------------------------------------------------------------------------
// Fechas específicas — UNA sección, un calendario
// ---------------------------------------------------------------------------
//
// Antes eran dos: "Días marcados" (multi-selección, sumaba días sueltos que SÍ
// se atienden) y "Excepciones por fecha" (selección simple, quitaba o alteraba
// un día). Vistas desde afuera son la misma pregunta con el signo cambiado —
// "en esta fecha pasa algo distinto de lo habitual" — y ambas se presentaban
// con un campo idéntico que decía "Elegir fecha". No había forma de saber cuál
// era cuál, ni por qué había dos.
//
// Ahora hay una: elegís la fecha y después decís qué pasa. El TIPO es lo que
// decide a qué callback del host va — "día extra" a OnDaysMarked, el resto a
// OnExceptionAdd — así que el contrato con el host no cambia; lo que cambia es
// que el usuario ve una decisión en vez de dos secciones sin nombre.
//
// Es el patrón "date overrides" de Calendly y Cal.com, y el equivalente en
// Acuity, Microsoft Bookings y Square: una lista de fechas que se apartan de
// la semana tipo, con un botón para agregar.

func (e *ScheduleEditor) buildExceptions() *Element {
	return Div().Set(clsExceptions.AsAttr()).
		Child(Div().Set(clsSectionTitle.AsAttr()).Text(lang.Translate("Specific dates").String())).
		Child(Div().Set(clsHintText.AsAttr()).
			Text(lang.Translate("Dates that differ from the weekly pattern").String())).
		Child(e.buildExceptionList()).
		Child(&calendarslider.CalendarSlider{
			NumMonths:  e.horizonMonths(),
			Holidays:   toCalHolidays(e.Holidays),
			Occupation: e.effectiveOccupation(),
			// Disponibilidad, no ocupación: acá 100 significa "ese día se
			// atiende". Sin decirlo, el widget lo pintaba con el rojo que
			// reserva para "no queda cupo" — rojo debajo de cada día que
			// trabajás.
			Meter:    calendarslider.Availability,
			Selected: e.sel,
			Expanded: e.calOpen,
			OnSelect: func(dateKey string) {
				e.sel.Set(dateKey)
				// calendarslider se pliega al elegir un día: correcto cuando
				// es un filtro —elegís y seguís— y lo contrario de lo que hace
				// falta acá, donde el calendario ES la pantalla y hay que ver
				// el efecto del cambio sobre el mes. Sin esto había que
				// reabrirlo a mano para comprobar si el día quedó marcado.
				e.calOpen.Set(true)
			},
		}).
		Child(e.buildExceptionForm())
}

func toCalHolidays(days []string) []calendarslider.Holiday {
	out := make([]calendarslider.Holiday, 0, len(days))
	for _, d := range days {
		out = append(out, calendarslider.Holiday{Date: d, Name: "Feriado"})
	}
	return out
}

// effectiveOccupation proyecta la agenda REAL sobre el calendario: el patrón
// semanal aplicado a cada fecha del horizonte, con las excepciones encima.
//
// Dos cosas a la vez, y por eso reemplaza a la lista de excepciones sueltas:
//
//  1. TODA fecha del horizonte queda seleccionable. calendarslider hace
//     seleccionable un día solo si aparece en Occupation, así que listar solo
//     las excepciones existentes dejaba el calendario muerto: no se podía
//     elegir una fecha que todavía no fuera excepción, que es lo único que se
//     quiere hacer ahí.
//  2. El calendario deja de ser un campo de entrada y pasa a ser el RESULTADO:
//     se ve el lunes/miércoles/viernes del patrón dibujado sobre los meses y,
//     encima, el día suelto que se agregó o se cerró. Cambiar el patrón cambia
//     el calendario, que es lo que convierte dos formularios independientes en
//     una sola afirmación sobre la agenda.
//
// El porcentaje es la barra que calendarslider ya pinta: 100 = ese día se
// atiende, 0 = no. No es "ocupación" en el sentido de reservas: es
// disponibilidad, que es lo que este editor decide.
func (e *ScheduleEditor) effectiveOccupation() []calendarslider.OccupationDay {
	months := e.horizonMonths()
	startYear, startMonth := date.ParseMonthKey(time.FormatDate(time.Now())[:7])
	if startYear == 0 {
		return nil
	}

	worksOn := [7]bool{}
	for _, row := range e.Pattern {
		for _, d := range row.Days {
			if d >= 0 && d <= 6 {
				worksOn[d] = true
			}
		}
	}

	out := make([]calendarslider.OccupationDay, 0, months*31)
	for m := 0; m < months; m++ {
		y, mon := date.AddMonths(startYear, startMonth, m)
		for d := 1; d <= date.DaysInMonth(y, mon); d++ {
			pct := 0
			if worksOn[date.Weekday(y, mon, d)] {
				pct = 100
			}
			out = append(out, calendarslider.OccupationDay{Date: date.DateKey(y, mon, d), Percent: pct})
		}
	}

	// Las excepciones mandan sobre el patrón: es lo que significa "excepción".
	for i := range out {
		for _, md := range e.Marked {
			if md.Date == out[i].Date {
				out[i].Percent = 100
			}
		}
		for _, ex := range e.Exceptions {
			if ex.Date != out[i].Date {
				continue
			}
			if ex.Type == ExcHoliday || ex.Type == ExcBlocked {
				out[i].Percent = 0
			} else {
				out[i].Percent = 100
			}
		}
	}
	return out
}

// ---------------------------------------------------------------------------
// La acción la elige el DÍA, no el usuario
// ---------------------------------------------------------------------------
//
// El formulario ofrecía cuatro tipos —Día extra / Cerrado / Horario especial /
// Bloqueado— sobre cualquier fecha. El problema no era que fueran muchos: era
// que la mayoría son INVÁLIDOS según el día que se tocó. Sobre un día que el
// patrón no cubre, "Cerrado" no hace nada (ya está cerrado) y "Horario
// especial" acota una ventana que no existe — el propio backend lo dice:
// "Abrir un día ya no necesita esta excepción" (service.go). De cuatro
// opciones, en un día no laborable sirve UNA.
//
// Así que el formulario no pedía una decisión: pedía esquivar trampas que el
// sistema ya sabe cuáles son. El calendario acaba de dibujar si ese día se
// trabaja; volver a preguntarlo es pedir que se repita lo que la pantalla
// mostró.
//
// Ahora el estado del día elige la acción y la acción pregunta a lo sumo una
// cosa más. El contrato con el host no cambia: se siguen emitiendo HOLIDAY,
// SPECIAL_HOURS y MarkedDay por los mismos callbacks.

// dayFacts es lo que se sabe de la fecha elegida, que es lo que decide qué
// acción ofrecer.
type dayFacts struct {
	covered bool // el patrón semanal cubre ese día de la semana
	exc     *Exception
	marked  *MarkedDay
}

func (e *ScheduleEditor) factsFor(key string) dayFacts {
	var f dayFacts
	y, m, d := date.ParseDateKey(key)
	if y != 0 {
		wd := date.Weekday(y, m, d)
		for _, row := range e.Pattern {
			if containsInt(row.Days, wd) {
				f.covered = true
				break
			}
		}
	}
	for i := range e.Exceptions {
		if e.Exceptions[i].Date == key {
			f.exc = &e.Exceptions[i]
			break
		}
	}
	for i := range e.Marked {
		if e.Marked[i].Date == key {
			f.marked = &e.Marked[i]
			break
		}
	}
	return f
}

func (e *ScheduleEditor) buildExceptionForm() *Element {
	form := Div().Set(clsExcForm.AsAttr()).
		BindStateFunc(widget.Open, func() bool { return e.sel.Get() != "" })

	// La fecha, legible. Antes decía "Fecha2026-09-02": sin separación y con
	// el identificador crudo que nadie lee.
	form.Child(Div().Set(clsExcDateRow.AsAttr()).
		Child(NewElement("time").
			Set(clsExcDate.AsAttr()).
			BindAttrFunc("datetime", func() string { return e.sel.Get() }).
			BindTextFunc(func() string { return humanDate(e.sel.Get()) })))

	form.Child(e.buildRevertAction())
	form.Child(e.buildCloseAction())
	form.Child(e.buildOpenAction())
	return form
}

// buildRevertAction: la fecha YA tiene una excepción. La única acción que tiene
// sentido es deshacerla — antes había que ir a buscarla a la lista, y tocar el
// día volvía a abrir el alta, de modo que se podían apilar dos excepciones
// sobre la misma fecha.
func (e *ScheduleEditor) buildRevertAction() *Element {
	box := Div().Set(clsExcAction.AsAttr()).
		BindStateFunc(widget.Open, func() bool {
			f := e.factsFor(e.sel.Get())
			return f.exc != nil || f.marked != nil
		})

	box.Child(Span().Set(clsExcHint.AsAttr()).
		BindTextFunc(func() string {
			f := e.factsFor(e.sel.Get())
			if f.marked != nil {
				return lang.Translate("Extra day").String()
			}
			if f.exc != nil {
				return exceptionLabel(f.exc.Type)
			}
			return ""
		}))

	revert := Button().Set(clsExcAdd.AsAttr()).
		Attr("type", "button").
		Text(lang.Translate("Back to normal hours").String())
	revert.OnClick(func(Event) {
		f := e.factsFor(e.sel.Get())
		if f.marked != nil && e.OnDaysUnmarked != nil {
			e.OnDaysUnmarked([]string{f.marked.Date})
		} else if f.exc != nil && e.OnExceptionRemove != nil {
			e.OnExceptionRemove(f.exc.ID)
		}
		e.sel.Set("")
	})
	return box.Child(revert)
}

// buildCloseAction: la fecha SÍ está en el patrón. Lo primario es cerrarla; lo
// secundario, atenderla en otro horario. Bloquear un rango a mitad del día no
// se ofrece acá: eso es tapar un hueco en una agenda ya definida, no definirla,
// y su etiqueta no lo distinguía de "horario especial" para nadie.
func (e *ScheduleEditor) buildCloseAction() *Element {
	box := Div().Set(clsExcAction.AsAttr()).
		BindStateFunc(widget.Open, func() bool {
			f := e.factsFor(e.sel.Get())
			return f.covered && f.exc == nil && f.marked == nil
		})

	notes := Input("text").Set(clsExcNotes.AsAttr()).
		Attr("name", "scheduleeditor-notes").
		Attr("placeholder", lang.Translate("Reason").String()).
		Bind(e.excNote)

	closeBtn := Button().Set(clsExcAdd.AsAttr()).
		Attr("type", "button").
		Text(lang.Translate("I do not work that day").String())
	closeBtn.OnClick(func(Event) {
		e.excType.Set(ExcHoliday)
		e.addException()
	})

	box.Child(notes).Child(closeBtn)

	box.Child(Span().Set(clsExcHint.AsAttr()).
		Text(lang.Translate("or work different hours that day").String()))
	box.Child(Div().Set(clsExcHours.AsAttr()).
		Child(boundTimePick(PartExcHours, "exc-from", e.excFrom, e.Bounds)).
		Child(boundTimePick(PartExcHours, "exc-to", e.excTo, e.Bounds)))

	special := Button().Set(clsExcAdd.AsAttr()).
		Attr("type", "button").
		Text(lang.Translate("Save these hours").String())
	special.OnClick(func(Event) {
		e.excType.Set(ExcSpecialHours)
		e.addException()
	})
	return box.Child(special)
}

// buildOpenAction: la fecha NO está en el patrón, así que lo único que tiene
// sentido es abrirla — un día suelto que se atiende fuera de la semana tipo.
func (e *ScheduleEditor) buildOpenAction() *Element {
	box := Div().Set(clsExcAction.AsAttr()).
		BindStateFunc(widget.Open, func() bool {
			f := e.factsFor(e.sel.Get())
			return !f.covered && f.exc == nil && f.marked == nil
		})

	box.Child(Span().Set(clsExcHint.AsAttr()).
		Text(lang.Translate("You do not work this weekday").String()))
	box.Child(Div().Set(clsExcHours.AsAttr()).
		Child(boundTimePick(PartExcHours, "exc-from", e.excFrom, e.Bounds)).
		Child(boundTimePick(PartExcHours, "exc-to", e.excTo, e.Bounds)))

	open := Button().Set(clsExcAdd.AsAttr()).
		Attr("type", "button").
		Text(lang.Translate("Work that day").String())
	open.OnClick(func(Event) {
		// SPECIAL_HOURS, no un "día extra": el backend resuelve esta excepción
		// ANTES de mirar los bloques semanales (service.go:564), así que abre
		// un día que el patrón no cubre. Y es la única vía que el cliente
		// expone — AddException existe, un guardado de bloques por fecha no.
		// El camino de MarkedDay quedaba en un callback que ningún host
		// cableaba, así que "Atender ese día" no hacía absolutamente nada.
		e.excType.Set(ExcSpecialHours)
		e.addException()
	})
	return box.Child(open)
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
	sel.OnChange(func(ev Event) {
		sig.Set(ev.TargetValue())
	})
	return sel
}

// dateEntry es una fila de la lista unificada: un día extra o una excepción.
// Existe para poder ordenar LAS DOS por fecha en una sola pasada — una lista
// que el usuario lee como una sola tiene que estar ordenada como una sola.
type dateEntry struct {
	Date   string
	Marked *MarkedDay
	Exc    *Exception
}

// buildExceptionList lista TODA fecha que se aparta de la semana tipo: los
// días extra (MarkedDay) y las excepciones (Exception), juntos y ordenados por
// fecha. Juntos porque el usuario no distingue "marcar" de "exceptuar" — ve
// una lista de fechas especiales — y porque separarlos es lo que dejaba a los
// días extra sin forma de quitarse cuando ambas secciones se fundieron.
func (e *ScheduleEditor) buildExceptionList() *Element {
	list := Ul().Set(clsExcList.AsAttr())

	if len(e.Exceptions) == 0 && len(e.Marked) == 0 && len(e.Holidays) == 0 {
		list.Child(Li().Set(clsExcItem.AsAttr()).
			Text(lang.Translate("No specific dates yet").String()))
		return list
	}

	for _, entry := range e.sortedDateEntries() {
		if entry.Marked != nil {
			list.Child(e.buildMarkedItem(*entry.Marked))
			continue
		}
		list.Child(e.buildExceptionItem(*entry.Exc))
	}
	return list
}

// sortedDateEntries funde ambas listas y las ordena por fecha (ISO, así que el
// orden lexicográfico ES el cronológico).
func (e *ScheduleEditor) sortedDateEntries() []dateEntry {
	out := make([]dateEntry, 0, len(e.Marked)+len(e.Exceptions))
	for i := range e.Marked {
		md := e.Marked[i]
		out = append(out, dateEntry{Date: md.Date, Marked: &md})
	}
	for _, ex := range sortedExceptions(e.Exceptions) {
		exc := ex
		out = append(out, dateEntry{Date: exc.Date, Exc: &exc})
	}
	for i := 0; i < len(out); i++ {
		min := i
		for j := i + 1; j < len(out); j++ {
			if out[j].Date < out[min].Date {
				min = j
			}
		}
		out[i], out[min] = out[min], out[i]
	}
	return out
}

func (e *ScheduleEditor) buildMarkedItem(md MarkedDay) *Element {
	row := Li().Set(clsExcItem.AsAttr()).Key("marked/" + md.Date)
	row.Child(dateCell(md.Date)).
		Child(Span().Text(lang.Translate("Extra day").String())).
		Child(Span().Text(hhmm(md.StartMin) + "–" + hhmm(md.EndMin)))

	remove := Button().Set(clsExcRemove.AsAttr()).
		Attr("type", "button").
		Text(lang.Translate("Remove").String())
	remove.OnClick(func(Event) {
		if e.OnDaysUnmarked != nil {
			e.OnDaysUnmarked([]string{md.Date})
		}
	})
	return row.Child(remove)
}

func (e *ScheduleEditor) buildExceptionItem(ex Exception) *Element {
	isHoliday := containsDate(e.Holidays, ex.Date)
	row := Li().Set(clsExcItem.AsAttr()).Key(ex.Date + "/" + ex.Type)

	row.Child(dateCell(ex.Date)).
		Child(Span().Text(exceptionLabel(ex.Type)))
	if hours, ok := exceptionHoursText(ex); ok {
		row.Child(Span().Text(hours))
	}
	if ex.Notes != "" {
		row.Child(Span().Set(clsExcNotes.AsAttr()).Text(ex.Notes))
	}

	// Un feriado lo pone el establecimiento, no el profesional: se muestra
	// pero no se puede quitar desde acá.
	if isHoliday {
		row.Set(clsExcHoliday.AsAttr())
		return row
	}
	remove := Button().Set(clsExcRemove.AsAttr()).
		Attr("type", "button").
		Text(lang.Translate("Remove").String())
	id := ex.ID
	remove.OnClick(func(Event) {
		if e.OnExceptionRemove != nil {
			e.OnExceptionRemove(id)
		}
	})
	return row.Child(remove)
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

// containsInt informa si el día d está en la lista de días de una PatternRow.
func containsInt(list []int, v int) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

// dateCell es la fecha de una fila: <time datetime="2026-09-17">17 de
// septiembre…</time>. El elemento que HTML tiene exactamente para esto, y que
// resuelve la tensión entera — la persona lee el texto, y la máquina (un
// lector de pantalla, un test, cualquier cosa que raspe la página) sigue
// teniendo el valor sin ambigüedad en el atributo.
func dateCell(key string) *Element {
	return NewElement("time").
		Set(clsExcDate.AsAttr()).
		Attr("datetime", key).
		Text(humanDate(key))
}

// humanDate escribe una fecha como se lee, no como se guarda. "2026-09-17" es
// un identificador: sirve para ordenar y comparar, y es ilegible en una lista
// que alguien recorre con la vista. La forma es la misma que usa el campo
// colapsado de calendarslider, para que la misma fecha se lea igual en las dos
// pantallas; las palabras las pone el diccionario de la app.
func humanDate(key string) string {
	y, m, d := date.ParseDateKey(key)
	if y == 0 {
		return key
	}
	weekday := date.WeekdayName(date.Weekday(y, m, d))
	return lang.Translate(weekday, d, date.MonthName(m), y).String()
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
		Child(e.buildExceptions())
}
