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

// shortWeekdayKeys son las claves canónicas en inglés del encabezado. La
// librería nunca fija un idioma ni el primer día: el texto lo traduce el
// diccionario de la app vía lang.Translate, y el orden lo decide
// date.FirstWeekday (lunes por defecto). Índice = convención de date.Weekday
// (0 = Sunday … 6 = Saturday).
var shortWeekdayKeys = [7]string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}

// maxMonths es el tope de meses navegables por slide, igual al límite del
// calendario original — evita una tira de scroll-snap sin control.
const maxMonths = 12

const iconCalendar = svg.Icon("cs-calendar")

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
	// actual (zona local).
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
	// SelectedMany holds space-separated date keys in multi-select mode.
	// When nil (default), single-select mode using Selected is active.
	SelectedMany *SignalString
	// Expanded controls whether the full calendar strip (true) or collapsed chip (false) is shown.
	Expanded *SignalBool
	// OnSelect se invoca al hacer clic en un día ocupable, con "YYYY-MM-DD".
	OnSelect func(date string)
	// OnToggle fires in multi-select mode with the date and its new state (true=selected, false=unselected).
	OnToggle func(date string, selected bool)

	onFilter func(term string) // set via OnFilterChange — satisfies widget.Filterable

	today string // fecha local de hoy, "YYYY-MM-DD"

	// months lets slideToMonth reach a month card's live node WITHOUT the
	// component inventing a global id — ids belong to dom.
	months []monthRef
}

// monthRef pairs a month key ("YYYY-MM") with the month card element.
type monthRef struct {
	key string
	el  *Element
}

var _ widget.Filterable = (*CalendarSlider)(nil)

func (c *CalendarSlider) OnFilterChange(fn func(term string)) { c.onFilter = fn }

func (c *CalendarSlider) holidayName(date string) (string, bool) {
	for _, h := range c.Holidays {
		if h.Date == date {
			return h.Name, true
		}
	}
	return "", false
}

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

func (c *CalendarSlider) Init(_ Ctx) {
	if c.Selected == nil {
		c.Selected = NewString("")
	}
	if c.Expanded == nil {
		c.Expanded = NewBool(true)
	}
	c.today = time.FormatDate(time.Now())
}

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

func (c *CalendarSlider) startYearMonth() (int, int) {
	sy, sm := date.ParseMonthKey(c.today)
	if c.Start != "" {
		if y, m := date.ParseMonthKey(c.Start); y != 0 {
			sy, sm = y, m
		}
	}
	return sy, sm
}

func (c *CalendarSlider) Render() *Element {
	n := c.numMonths()
	sy, sm := c.startYearMonth()

	keys := make([]string, n)
	for i := 0; i < n; i++ {
		y, m := date.AddMonths(sy, sm, i)
		keys[i] = date.MonthKey(y, m)
	}

	c.months = c.months[:0]

	strip := Div().Set(clsStrip.AsAttr()).
		Attr("role", "grid").
		Attr("aria-label", "Calendario").
		BindState(widget.Current, c.Expanded)
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

func (c *CalendarSlider) buildCollapsed() *Element {
	toggle := Input("checkbox").Set(clsCollapsedToggle.AsAttr()).
		BindAttrBool("checked", c.Expanded).
		On("change", func(e Event) {
			expanded := e.TargetChecked()
			c.Expanded.Set(expanded)
			if !expanded && c.SelectedMany == nil && c.Selected.Get() == "" && c.today != "" {
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
			selKey := ""
			if c.SelectedMany != nil {
				selKey = c.SelectedMany.Get()
			} else if c.Selected != nil {
				selKey = c.Selected.Get()
			}
			y, m, d := date.ParseDateKey(selKey)
			if y == 0 {
				return ""
			}
			weekday := date.WeekdayName(date.Weekday(y, m, d))
			return lang.Translate(weekday, d, date.MonthName(m), y).String()
		})

	label := Label().Set(clsCollapsed.AsAttr()).
		Child(toggle).
		Child(Div().Set(clsCollapsedCap.AsAttr()).
			Child(iconCalendar.Render(string(clsCollapsedIcon)))).
		Child(text)

	return Div().Child(label)
}

func (c *CalendarSlider) buildWeekdayRow() *Element {
	weekdays := Ul().Set(clsWeekRow.AsAttr()).Attr("role", "row")
	for _, w := range date.WeekOrder() {
		weekdays.Child(Li().Set(clsWeekday.AsAttr()).
			Attr("role", "columnheader").
			Text(lang.Translate(shortWeekdayKeys[w]).String()))
	}
	return weekdays
}

const weeksPerMonth = 6

func (c *CalendarSlider) buildMonth(year, month int, prevKey, nextKey string, prevWraps, nextWraps bool) *Element {
	key := date.MonthKey(year, month)
	monthEl := Div().Set(clsMonth.AsAttr()).
		Key(key).
		Attr("data-month", key)

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
		Attr("data-target", prevKey).
		Text("‹")
	prev.On("click", func(Event) { c.slideToMonth(prevKey, prevWraps) })
	next := Button().Set(clsNext.AsAttr()).
		Attr("type", "button").
		Attr("aria-label", "Mes siguiente").
		Attr("title", "Mes siguiente").
		Attr("data-target", nextKey).
		Text("›")
	next.On("click", func(Event) { c.slideToMonth(nextKey, nextWraps) })

	monthEl.Child(Div().Set(clsMonthNav.AsAttr()).
		Child(prev).
		Child(monthName).
		Child(next))

	c.months = append(c.months, monthRef{key: key, el: monthEl})

	return monthEl
}

func (c *CalendarSlider) slideToMonth(key string, instant bool) {
	var el *Element
	for i := range c.months {
		if c.months[i].key == key {
			el = c.months[i].el
			break
		}
	}
	if el == nil {
		return
	}
	ref, ok := el.Ref()
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
	startCol := date.WeekColumn(date.Weekday(year, month, 1))
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
	selectable := use >= 0 && holiday == ""

	title := ""
	switch {
	case isToday:
		title = "Hoy"
	case holiday != "":
		title = holiday
	case selectable:
		title = fmt.Sprint(use) + "%"
	}

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

	isSel := DeriveBool(func() bool {
		if c.SelectedMany != nil {
			return containsWord(c.SelectedMany.Get(), dateStr)
		}
		if c.Selected != nil {
			return c.Selected.Get() == dateStr
		}
		return false
	})

	stack := Div().Set(clsDayStack.AsAttr()).
		Child(Span().Set(clsDayNum.AsAttr()).Text(fmt.Sprint(day)))
	if selectable {
		stack.Child(Div().Set(clsDayUse.AsAttr()).
			Attr("data-use", fmt.Sprint(use)).
			Attr("style", "--meter-fill:"+fmt.Sprint(use)+"%;"))
	}

	li := Li().Set(classes...).
		Key(dateStr).
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
			if c.SelectedMany != nil {
				newVal, isNowSelected := toggleWord(c.SelectedMany.Get(), dateStr)
				c.SelectedMany.Set(newVal)
				if c.OnToggle != nil {
					c.OnToggle(dateStr, isNowSelected)
				}
			} else {
				if c.Selected != nil {
					c.Selected.Set(dateStr)
				}
				c.Expanded.Set(false)
				if c.OnSelect != nil {
					c.OnSelect(dateStr)
				}
				if c.onFilter != nil {
					c.onFilter(dateStr)
				}
			}
		})
	}
	return li
}

func containsWord(s, word string) bool {
	wLen := len(word)
	sLen := len(s)
	if wLen == 0 || sLen < wLen {
		return false
	}
	for i := 0; i <= sLen-wLen; i++ {
		if (i == 0 || s[i-1] == ' ' || s[i-1] == ',') && (i+wLen == sLen || s[i+wLen] == ' ' || s[i+wLen] == ',') {
			if s[i:i+wLen] == word {
				return true
			}
		}
	}
	return false
}

func toggleWord(s, word string) (string, bool) {
	if containsWord(s, word) {
		out := ""
		wLen := len(word)
		sLen := len(s)
		for i := 0; i < sLen; {
			if (i == 0 || s[i-1] == ' ' || s[i-1] == ',') && (i+wLen <= sLen && s[i:i+wLen] == word) && (i+wLen == sLen || s[i+wLen] == ' ' || s[i+wLen] == ',') {
				i += wLen
				if i < sLen && (s[i] == ' ' || s[i] == ',') {
					i++
				}
			} else {
				out += string(s[i])
				i++
			}
		}
		if len(out) > 0 && (out[len(out)-1] == ' ' || out[len(out)-1] == ',') {
			out = out[:len(out)-1]
		}
		return out, false
	}
	if s == "" {
		return word, true
	}
	return s + " " + word, true
}
