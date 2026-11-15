package agedate

import (
	"webtyp.com/dom"
	"webtyp.com/fmt"

	. "webtyp.com/html"
	"webtyp.com/input"
	"webtyp.com/lang"
	"webtyp.com/time"
	"webtyp.com/widget"
)

const NameAgeDate = widget.Name("agedate")
const (
	PartAge   = widget.Part("age")
	PartDate  = widget.Part("date")
	PartError = widget.Part("error") // why the typed age was not taken
)

// Messages for an age that cannot become a birth date.
const (
	textUnderOneYear lang.Text = "Under one year needs the exact birth date"
	textWholeYears   lang.Text = "Enter the age in whole years"
)

// ARIA/attribute values for the "also sets" checkbox the model names.
const (
	setEstimated = "true"
	setExact     = "false"
)

// AgeDate is a birth-date input that also accepts the age the person says.
// Typing an age (whole years, 1 or more) fills the date and, when the model
// names a field in Field.AlsoSetsField, marks it as estimated; typing the exact
// date marks it as not estimated. Under one year there is no age: the exact
// date is required. It implements input.AlsoSetter.
type AgeDate struct {
	input.Base
	date input.Input
	set  func(string) // writes the field named by Field.AlsoSetsField; nil until the form binds it
}

// BindAlsoSets receives the function that writes the field Field.AlsoSetsField
// names (see input.AlsoSetter).
func (a *AgeDate) BindAlsoSets(set func(string)) { a.set = set }

func (a *AgeDate) alsoSet(value string) {
	if a.set != nil {
		a.set(value)
	}
}

func New() *AgeDate {
	a := &AgeDate{
		date: input.Date(),
	}
	a.InitBase("", "", "date")
	return a
}

func (a *AgeDate) Clone(parentID, name string) input.Input {
	c := New()
	c.InitBase(parentID, name, "date")
	c.set = a.set
	return c
}

func (a *AgeDate) Validate(value string) error {
	if value == "" {
		return nil
	}
	return a.date.Validate(value)
}

func (a *AgeDate) WidgetName() widget.Name { return NameAgeDate }
func (a *AgeDate) WidgetKind() widget.Kind { return widget.Combobox }

// today is the date on the clock of the computer at the front desk (the clinic's): the age is counted from it.
// MachineZone is right here: whoever types the age stands in front of that computer.
func today() string { return time.MachineZone().Date(time.NowUnix()) }

// ageYears reads a typed age: whole years, 1 or more. An empty text is not a
// problem (nothing typed); anything else that is not that names the problem.
func ageYears(ageText string) (years int, problem lang.Text) {
	text := fmt.Convert(ageText).TrimSpace().String()
	if text == "" {
		return 0, ""
	}
	if _, err := fmt.Convert(text).Replace(",", ".").Float64(); err != nil {
		return 0, textWholeYears
	}
	val, err := fmt.Convert(text).Int()
	if err != nil {
		if f, ferr := fmt.Convert(text).Replace(",", ".").Float64(); ferr == nil && f > 0 && f < 1 {
			return 0, textUnderOneYear
		}
		return 0, textWholeYears
	}
	if val < 1 {
		return 0, textUnderOneYear
	}
	return int(val), ""
}

// birthFromAge: today ("YYYY-MM-DD") minus whole years, same month and day.
// "", "0", a fraction or a negative → "", false (see ageYears for the reason).
func birthFromAge(ageText, today string) (date string, ok bool) {
	years, problem := ageYears(ageText)
	if problem != "" || years < 1 {
		return "", false
	}
	t, err := time.ParseDateUTC(today)
	if err != nil {
		return "", false
	}
	ty, tm, td, _, _, _ := time.CivilUTC(t)
	y, m, d, _, _, _ := time.CivilUTC(time.DateUTC(ty-years, tm, td, 0, 0, 0))
	return fmt.Sprintf("%04d-%02d-%02d", y, m, d), true
}

// ageFromBirth: whole years on today ("YYYY-MM-DD"). A date less than a year
// old has no age to show: "", false.
func ageFromBirth(date, today string) (age string, ok bool) {
	b, err := time.ParseDateUTC(date)
	if err != nil {
		return "", false
	}
	t, err := time.ParseDateUTC(today)
	if err != nil {
		return "", false
	}
	ty, tm, td, _, _, _ := time.CivilUTC(t)
	by, bm, bd, _, _, _ := time.CivilUTC(b)
	years := ty - by
	if tm < bm || (tm == bm && td < bd) {
		years--
	}
	if years < 1 {
		return "", false
	}
	return fmt.Sprintf("%d", years), true
}

// RenderInput draws the age box and the date box. The form calls onInput when
// a value is committed (the box loses focus or the date is picked), never per
// keystroke; the age is not a value of the record, so a bad age is explained
// right under it and nothing is sent.
func (a *AgeDate) RenderInput(value *dom.SignalString, onInput func(string)) *dom.Element {
	ageVal := ""
	if v := value.Get(); v != "" {
		if age, ok := ageFromBirth(v, today()); ok {
			ageVal = age
		}
	}
	problem := dom.NewString("")
	commit := func(v string) {
		if onInput != nil {
			onInput(v)
		}
	}

	ageInput := Input("number").
		Attr("min", "1").Attr("max", "130").Attr("step", "1").
		Attr("value", ageVal).
		Class(NameAgeDate.Class(PartAge).String()).
		OnChange(func(e dom.Event) {
			text := e.TargetValue()
			if _, p := ageYears(text); p != "" {
				problem.Set(lang.Translate(p).String())
				return
			}
			problem.Set("")
			date, ok := birthFromAge(text, today())
			if !ok {
				return
			}
			value.Set(date)
			a.alsoSet(setEstimated)
			commit(date)
		})

	dateInput := Input("date").
		Attr("value", value.Get()).
		Class(NameAgeDate.Class(PartDate).String()).
		OnChange(func(e dom.Event) {
			val := e.TargetValue()
			value.Set(val)
			problem.Set("")
			age, _ := ageFromBirth(val, today())
			if ref, ok := ageInput.Ref(); ok {
				ref.SetValue(age)
			}
			a.alsoSet(setExact)
			commit(val)
		}).
		Bind(value)

	ageId := "age-" + a.GetID()
	ageInput.ID(ageId)

	row := Div().Class(NameAgeDate.Root().String()).Child(
		Label().For(ageInput).Text(lang.Translate("Age").String()),
		ageInput,
		dateInput,
	)
	return Div().Child(
		row,
		dom.NewElement("small").Class(NameAgeDate.Class(PartError).String()).
			Attr("role", "alert").BindText(problem),
	)
}
