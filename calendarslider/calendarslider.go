// Package calendarslider ports the legacy "calendar normal" widget — a
// single-month view with holidays, occupation percentages and day selection,
// sliding between neighboring months — to the webtyp construction harness:
// pure Go, zero JavaScript. The old infinite slider (JS that animated
// margin-left and recycled DOM nodes) becomes a bounded, pre-rendered strip
// of up to maxMonths months. The ‹ › controls are <button>s; a small click
// handler (slideToMonth) jumps the scroll-snap strip to the neighbouring
// month card with ScrollIntoView, and the browser's scroll-snap animates it.
// They are deliberately not <a href="#cs-m-...">: an anchor mutates
// location.hash, which a hash-routed shell (platformd) reads as a route
// change, blanking the view.
package calendarslider

import (
	"webtyp.com/date"
	. "webtyp.com/dom"
	"webtyp.com/fmt"
	"webtyp.com/fmt/lang"
	. "webtyp.com/html"
	"webtyp.com/svg"
	"webtyp.com/time"
	"webtyp.com/widget"
)

// NameCalendarSlider is the widget identity.
const NameCalendarSlider = widget.Name("calendarslider")

const (
	PartWeekday         = widget.Part("weekday")
	PartWeekRow         = widget.Part("week-row")
	PartStrip           = widget.Part("strip")
	PartMonth           = widget.Part("month")
	PartMonthName       = widget.Part("month-name")
	PartDay             = widget.Part("day")
	PartDayNum          = widget.Part("day-num")
	PartDayStack        = widget.Part("day-stack")
	PartDayUse          = widget.Part("day-use")
	PartDaySelectable   = widget.Part("day-selectable")
	PartDayOff          = widget.Part("day-off")
	PartDayRed          = widget.Part("day-red")
	PartDayToday        = widget.Part("day-today")
	PartPrev            = widget.Part("prev")
	PartNext            = widget.Part("next")
	PartMonthNav        = widget.Part("month-nav")
	PartCollapsed       = widget.Part("collapsed")
	PartCollapsedToggle = widget.Part("collapsed-toggle")
	PartCollapsedCap    = widget.Part("collapsed-cap")
	PartCollapsedIcon   = widget.Part("collapsed-icon")
	PartCollapsedText   = widget.Part("collapsed-text")
)

var (
	clsRoot            = NameCalendarSlider.Root()
	clsWeekday         = NameCalendarSlider.Class(PartWeekday)
	clsWeekRow         = NameCalendarSlider.Class(PartWeekRow)
	clsStrip           = NameCalendarSlider.Class(PartStrip)
	clsMonth           = NameCalendarSlider.Class(PartMonth)
	clsMonthNm         = NameCalendarSlider.Class(PartMonthName)
	clsDay             = NameCalendarSlider.Class(PartDay)
	clsDayNum          = NameCalendarSlider.Class(PartDayNum)
	clsDayStack        = NameCalendarSlider.Class(PartDayStack)
	clsDayUse          = NameCalendarSlider.Class(PartDayUse)
	clsDaySel          = NameCalendarSlider.Class(PartDaySelectable)
	clsDayOff          = NameCalendarSlider.Class(PartDayOff)
	clsDayRed          = NameCalendarSlider.Class(PartDayRed)
	clsDayToday        = NameCalendarSlider.Class(PartDayToday)
	clsPrev            = NameCalendarSlider.Class(PartPrev)
	clsNext            = NameCalendarSlider.Class(PartNext)
	clsMonthNav        = NameCalendarSlider.Class(PartMonthNav)
	clsCollapsed       = NameCalendarSlider.Class(PartCollapsed)
	clsCollapsedToggle = NameCalendarSlider.Class(PartCollapsedToggle)
	clsCollapsedCap    = NameCalendarSlider.Class(PartCollapsedCap)
	clsCollapsedIcon   = NameCalendarSlider.Class(PartCollapsedIcon)
	clsCollapsedText   = NameCalendarSlider.Class(PartCollapsedText)
)

// weekdayNames es el encabezado de cada sección: la semana comienza en lunes,
// igual que la implementación original.
var weekdayNames = [7]string{"Lun", "Mar", "Mie", "Jue", "Vie", "Sab", "Dom"}

// maxMonths es el tope de meses navegables por slide, igual al límite del
// calendario original — evita una tira de scroll-snap sin control.
const maxMonths = 12

const iconCalendar = svg.Icon("cs-calendar")

var calendarSliderSeq int

func nextCalendarSliderID() int {
	calendarSliderSeq++
	return calendarSliderSeq
}

const suffixCollapsedToggle = "-collapsed-toggle"

// Holiday es un feriado del calendario.
type Holiday struct {
	Date string // "YYYY-MM-DD"
	Name string
}

// OccupationDay es el porcentaje de ocupación (0..100) de una fecha; su sola
// presencia en la lista hace el día seleccionable.
type OccupationDay struct {
	Date    string // "YYYY-MM-DD"
	Percent int
}

// CalendarSlider muestra un mes a la vez, empezando en Start, con feriados,
// porcentaje de ocupación, marcador de hoy y selección de día; ‹ › deslizan
// hacia los meses vecinos. Los días con ocupación son los únicos
// seleccionables — el resto se muestra como día inactivo, igual que el
// calendario "normal" original.
type CalendarSlider struct {
	Element // value embed — NEVER pointer (TinyGo heap constraint)

	// Start es el primer mes de la tira, formato "YYYY-MM"; vacío = el mes
	// actual (zona local). Hacia atrás de Start no hay nada que deslizar —
	// igual que el calendario original, pensado para reservar hacia
	// adelante, no para consultar meses pasados.
	Start string
	// NumMonths es cuántos meses hay para deslizar hacia adelante desde
	// Start; 0 = 3, tope maxMonths (12).
	NumMonths int
	// Holidays lista los feriados del calendario. Slice, no map — TinyGo.
	Holidays []Holiday
	// Occupation lista el porcentaje de ocupación por fecha. Slice, no map —
	// TinyGo.
	Occupation []OccupationDay
	// Selected es la fecha "YYYY-MM-DD" seleccionada, o "". Señal pública:
	// el host puede leerla y escribirla.
	Selected *SignalString
	// OnSelect se invoca al hacer clic en un día ocupable, con "YYYY-MM-DD".
	OnSelect func(date string)

	onFilter func(term string) // set via OnFilterChange — satisfies widget.Filterable

	today string // fecha local de hoy, "YYYY-MM-DD"

	uid      string      // per-instance id prefix; two calendars on one page must not collide
	expanded *SignalBool // true = full calendar showing; false = mobile collapsed chip showing
}

var _ widget.Filterable = (*CalendarSlider)(nil)

// OnFilterChange implements widget.Filterable: it registers the sink called
// with the picked day ("YYYY-MM-DD") on every day selection. The signature
// is fixed by widget.Filterable — do not add a parameter, do not rename.
func (c *CalendarSlider) OnFilterChange(fn func(term string)) { c.onFilter = fn }

// holidayName busca date en Holidays. Recorrido lineal — a lo sumo unas
// pocas decenas de entradas por ventana renderizada — no map, TinyGo.
func (c *CalendarSlider) holidayName(date string) (string, bool) {
	for _, h := range c.Holidays {
		if h.Date == date {
			return h.Name, true
		}
	}
	return "", false
}

// occupationPercent busca date en Occupation. Recorrido lineal, misma razón
// que holidayName.
func (c *CalendarSlider) occupationPercent(date string) (int, bool) {
	for _, o := range c.Occupation {
		if o.Date == date {
			return o.Percent, true
		}
	}
	return 0, false
}

func (c *CalendarSlider) WidgetName() widget.Name { return NameCalendarSlider }
func (c *CalendarSlider) WidgetKind() widget.Kind { return widget.Grid }

func (c *CalendarSlider) ensureUID() {
	if c.uid == "" {
		c.uid = fmt.Sprintf("%s-%d", string(NameCalendarSlider), nextCalendarSliderID())
	}
}

func (c *CalendarSlider) Init(_ Ctx) {
	c.ensureUID()
	if c.Selected == nil {
		c.Selected = NewString("")
	}
	if c.expanded == nil {
		c.expanded = NewBool(true)
	}
	c.today = time.FormatDate(time.Now())
}

// numMonths aplica el default (0 = 3) y el tope maxMonths.
func (c *CalendarSlider) numMonths() int {
	n := c.NumMonths
	if n < 1 {
		n = 3
	}
	if n > maxMonths {
		n = maxMonths
	}
	return n
}

// startYearMonth resuelve el (año, mes) inicial de la tira: Start si es
// válido, si no el mes de hoy.
func (c *CalendarSlider) startYearMonth() (int, int) {
	sy, sm := date.ParseMonthKey(c.today)
	if c.Start != "" {
		if y, m := date.ParseMonthKey(c.Start); y != 0 {
			sy, sm = y, m
		}
	}
	return sy, sm
}

// Render arma la tira completa de meses de una sola vez, Start primero —
// sin señal, sin reconstrucción: el slide entre ellos lo hace el
// scroll-snap del navegador, disparado por los enlaces ‹ › de cada mes hacia
// el vecino. Al no reconstruirse nunca, el mes inicial es siempre el primer
// hijo de la tira — la posición de scroll 0 — sin necesitar desplazar el
// scroll por WASM al montar.
//
// La navegación es un bucle: el ‹ del primer mes apunta al último y el › del
// último apunta al primero, igual que el deslizador infinito original — la
// alternativa (sin bucle) obliga a recorrer los N meses en orden para volver
// al principio.
func (c *CalendarSlider) Render() *Element {
	c.ensureUID()
	n := c.numMonths()
	sy, sm := c.startYearMonth()

	keys := make([]string, n)
	for i := 0; i < n; i++ {
		y, m := date.AddMonths(sy, sm, i)
		keys[i] = date.MonthKey(y, m)
	}

	strip := Div().Set(clsStrip.AsAttr()).
		Attr("role", "grid").
		Attr("aria-label", "Calendario").
		BindState(widget.Current, c.expanded)
	for i, key := range keys {
		y, m := date.ParseMonthKey(key)
		prevKey := keys[(i-1+n)%n]
		nextKey := keys[(i+1)%n]
		prevWraps := i == 0
		nextWraps := i == n-1
		strip.Child(c.buildMonth(y, m, prevKey, nextKey, prevWraps, nextWraps))
	}

	return Div().Set(clsRoot.AsAttr()).
		Child(strip).
		Child(c.buildCollapsed())
}

// buildCollapsed builds the bottom bar the calendar folds into and unfolds
// from: a hidden checkbox + <label> — selectsearch's own open/close idiom
// (see its PartHeader/toggle, selectsearch.go ~213-261) — toggles
// c.expanded with the same two-way sync, no JS beyond that one listener.
// Rendered last, full width, in normal flow under the strip, on every
// viewport (it is the fold control even while expanded: tap to collapse,
// tap to expand) — see the collapse rules in css.go. Both strip and chip
// read the same c.expanded signal in opposite directions (strip follows it,
// the chip's checkbox mirrors it); the chip carries no state attribute of
// its own because no rule selects on it.
//
// The date text goes through lang.Translate, exactly like the month label
// and for the identical reason: date.WeekdayName/date.MonthName return
// English canonical names, never a hardcoded language — this component
// registers no dictionary itself, it only asks to translate. With no
// dictionary registered anywhere in the running app this reads "Monday 18
// August 2026"; a host that has registered Spanish sees "Lunes 18 Agosto
// 2026" with no further code change here.
func (c *CalendarSlider) buildCollapsed() *Element {
	c.ensureUID()
	toggle := Input("checkbox").Set(clsCollapsedToggle.AsAttr()).
		ID(c.uid+suffixCollapsedToggle).
		BindAttrBool("checked", c.expanded).
		On("change", func(e Event) {
			expanded := e.TargetChecked()
			c.expanded.Set(expanded)
			// Folding with no day picked picks today: the chip must never
			// fold onto a blank label. Identical to tapping today itself —
			// same signal, same callbacks — so the list filters to today.
			if !expanded && c.Selected.Get() == "" && c.today != "" {
				c.Selected.Set(c.today)
				if c.OnSelect != nil {
					c.OnSelect(c.today)
				}
				if c.onFilter != nil {
					c.onFilter(c.today)
				}
			}
		})

	text := Span().Set(clsCollapsedText.AsAttr()).
		BindTextFunc(func() string {
			y, m, d := date.ParseDateKey(c.Selected.Get())
			if y == 0 {
				return ""
			}
			weekday := date.WeekdayName(date.Weekday(y, m, d))
			return lang.Translate(weekday, d, date.MonthName(m), y).String()
		})

	label := Label().Set(clsCollapsed.AsAttr()).
		Attr("for", c.uid+suffixCollapsedToggle).
		Child(Div().Set(clsCollapsedCap.AsAttr()).
			Child(iconCalendar.Render(string(clsCollapsedIcon)))).
		Child(text)

	return Div().Child(toggle).Child(label)
}

// buildWeekdayRow arma la fila de días de la semana (lunes primero) que cada
// sección de mes lleva en su cabecera, igual que el original.
func (c *CalendarSlider) buildWeekdayRow() *Element {
	weekdays := Ul().Set(clsWeekRow.AsAttr()).Attr("role", "row")
	for _, name := range weekdayNames {
		weekdays.Child(Li().Set(clsWeekday.AsAttr()).
			Attr("role", "columnheader").
			Text(name))
	}
	return weekdays
}

// weeksPerMonth es el número de filas de semana que todo mes reserva,
// ocupe o no sus 7 días — el máximo real (un mes de 31 días que empieza en
// domingo cae en 6). Sin este piso fijo, un mes de 4 o 5 filas deja una
// tarjeta más baja y la etiqueta del mes (y el ‹ › que se ancla a la
// tarjeta) saltan verticalmente al deslizar entre meses.
const weeksPerMonth = 6

// buildMonth arma un mes: fila de días de la semana (lunes primero), 6 filas
// de semana siempre (rellenas con celdas vacías si el mes real tiene menos,
// ver weeksPerMonth), la etiqueta del mes debajo (como el original,
// footer-title-month) y sus propios enlaces ‹ › hacia los meses vecinos —
// siempre los dos, el primero y el último de la tira se enlazan entre sí
// (ver Render). Cada mes es el único visible a la vez (scroll-snap en
// PartStrip), así que sus flechas son, en la práctica, "las" flechas de
// navegación mientras esté en pantalla.
func (c *CalendarSlider) buildMonth(year, month int, prevKey, nextKey string, prevWraps, nextWraps bool) *Element {
	c.ensureUID()
	key := date.MonthKey(year, month)
	monthEl := Div().Set(clsMonth.AsAttr()).
		Key(key).
		ID(c.uid + "-m-" + key)

	monthEl.Child(c.buildWeekdayRow())

	cells := c.buildDayCells(year, month)
	weeks := 0
	for start := 0; start < len(cells); start += 7 {
		end := start + 7
		if end > len(cells) {
			end = len(cells)
		}
		week := Ul().Set(clsWeekRow.AsAttr()).Attr("role", "row")
		for _, cell := range cells[start:end] {
			week.Child(cell)
		}
		monthEl.Child(week)
		weeks++
	}
	for ; weeks < weeksPerMonth; weeks++ {
		// clsDay (sin variante day-off/day-selectable/etc.) para que la fila
		// de relleno mida lo mismo que una fila real — un <li> vacío sin la
		// caja de IconBox colapsa a la altura de una línea de texto y la
		// tarjeta vuelve a variar de alto entre meses.
		fillerWeek := Ul().Set(clsWeekRow.AsAttr()).Attr("role", "row").Attr("aria-hidden", "true")
		for i := 0; i < 7; i++ {
			fillerWeek.Child(Li().Set(clsDay.AsAttr()).Attr("aria-hidden", "true"))
		}
		monthEl.Child(fillerWeek)
	}

	monthName := Div().Set(clsMonthNm.AsAttr()).
		Text(lang.Translate(date.MonthName(month), year).String())

	prev := Button().Set(clsPrev.AsAttr()).
		Attr("type", "button").
		Attr("aria-label", "Mes anterior").
		Attr("title", "Mes anterior").
		Attr("data-target", c.uid+"-m-"+prevKey).
		Text("‹")
	prev.On("click", func(Event) { c.slideToMonth(prevKey, prevWraps) })
	next := Button().Set(clsNext.AsAttr()).
		Attr("type", "button").
		Attr("aria-label", "Mes siguiente").
		Attr("title", "Mes siguiente").
		Attr("data-target", c.uid+"-m-"+nextKey).
		Text("›")
	next.On("click", func(Event) { c.slideToMonth(nextKey, nextWraps) })

	monthEl.Child(Div().Set(clsMonthNav.AsAttr()).
		Child(prev).
		Child(monthName).
		Child(next))

	return monthEl
}

// slideToMonth jumps the scroll-snap strip to the month card carrying the
// given key. instant selects ScrollIntoViewInstant over the normal smooth
// ScrollIntoView — reserved for the two wrap edges (first month's ‹, last
// month's ›), where a smooth scroll would visibly travel across every
// month in between in the wrong apparent direction. Every adjacent-month
// navigation keeps calling this with instant=false.
func (c *CalendarSlider) slideToMonth(key string, instant bool) {
	ref, ok := Get(c.uid + "-m-" + key)
	if !ok {
		return
	}
	if instant {
		ref.ScrollIntoViewInstant()
		return
	}
	ref.ScrollIntoView()
}

func (c *CalendarSlider) buildDayCells(year, month int) []*Element {
	startCol := (date.Weekday(year, month, 1) + 6) % 7 // 0 = lunes
	total := date.DaysInMonth(year, month)
	cells := make([]*Element, 0, startCol+total)
	for i := 0; i < startCol; i++ {
		cells = append(cells, Li().Attr("aria-hidden", "true"))
	}
	for day := 1; day <= total; day++ {
		cells = append(cells, c.buildDay(year, month, day))
	}
	return cells
}

func (c *CalendarSlider) buildDay(year, month, day int) *Element {
	c.ensureUID()
	dateStr := date.DateKey(year, month, day)
	weekday := date.Weekday(year, month, day)

	use := -1
	if v, ok := c.occupationPercent(dateStr); ok {
		use = v
		if use < 0 {
			use = 0
		}
		if use > 100 {
			use = 100
		}
	}
	holiday := ""
	if name, ok := c.holidayName(dateStr); ok {
		holiday = name
	}

	isToday := c.today == dateStr
	selectable := use >= 0

	title := ""
	switch {
	case isToday:
		title = "Hoy"
	case holiday != "":
		title = holiday
	case selectable:
		title = fmt.Sprint(use) + "%"
	}

	// Un día con ocupación gana sobre domingo/feriado, igual que en el
	// original (el bloque de ocupación reemplazaba la clase roja); hoy es
	// aditivo por encima de cualquier variante.
	classes := []fmt.KeyValue{clsDay.AsAttr()}
	switch {
	case selectable:
		classes = append(classes, clsDaySel.AsAttr())
	case weekday == 0 || holiday != "":
		classes = append(classes, clsDayRed.AsAttr())
	default:
		classes = append(classes, clsDayOff.AsAttr())
	}
	if isToday {
		classes = append(classes, clsDayToday.AsAttr())
	}

	isSel := DeriveBool(func() bool { return c.Selected.Get() == dateStr })

	stack := Div().Set(clsDayStack.AsAttr()).
		Child(Span().Set(clsDayNum.AsAttr()).Text(fmt.Sprint(day)))
	if selectable {
		stack.Child(Div().Set(clsDayUse.AsAttr()).
			Attr("data-use", fmt.Sprint(use)).
			Attr("style", "--meter-fill:"+fmt.Sprint(use)+"%;"))
	}

	li := Li().Set(classes...).
		Key(dateStr).
		ID(c.uid + "-d-" + dateStr).
		Attr("role", "gridcell").
		Attr("data-date", dateStr).
		BindState(widget.Selected, isSel).
		BindAttrFunc("aria-selected", func() string {
			if isSel.Get() {
				return "true"
			}
			return "false"
		}).
		Child(stack)
	if title != "" {
		li.Attr("title", title)
	}
	if selectable {
		li.On("click", func(Event) {
			c.Selected.Set(dateStr)
			c.expanded.Set(false)
			if c.OnSelect != nil {
				c.OnSelect(dateStr)
			}
			if c.onFilter != nil {
				c.onFilter(dateStr)
			}
		})
	}
	return li
}
