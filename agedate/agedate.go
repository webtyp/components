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
	PartAge  = widget.Part("age")
	PartDate = widget.Part("date")
)

type AgeDate struct {
	input.Base
	date input.Input
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

// birthFromAge: today ("YYYY-MM-DD") minus years and months. ageText uses '.' or ',' as decimal mark; the
// decimal part is a fraction of a year (0.5 = 6 months). "" or "0" → "", false.
func birthFromAge(ageText, today string) (date string, ok bool) {
	val, err := fmt.Convert(ageText).Replace(",", ".").Float64()
	if err != nil || val <= 0 {
		return "", false
	}
	years := int(val)
	months := int((val - float64(years)) * 12)

	t, err := time.ParseDateUTC(today)
	if err != nil {
		return "", false
	}
	ty, tm, td, _, _, _ := time.CivilUTC(t)
	newYear, newMonth := ty-years, tm-months
	for newMonth <= 0 {
		newMonth += 12
		newYear--
	}
	y, m, d, _, _, _ := time.CivilUTC(time.DateUTC(newYear, newMonth, td, 0, 0, 0))
	return fmt.Sprintf("%04d-%02d-%02d", y, m, d), true
}

// ageFromBirth: whole years, plus the year fraction of full months, rounded to 1 decimal ("40.3"), counted on today
// ("YYYY-MM-DD").
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
	years, months := ty-by, tm-bm
	if td < bd {
		months--
	}
	if months < 0 {
		years--
		months += 12
	}
	if years < 0 {
		return "", false
	}
	return fmt.Sprintf("%.1f", float64(years)+float64(months)/12.0), true
}

func (a *AgeDate) RenderInput(value *dom.SignalString, onInput func(string)) *dom.Element {
	ageVal := ""
	if v := value.Get(); v != "" {
		if age, ok := ageFromBirth(v, today()); ok {
			ageVal = age
		}
	}

	ageInput := Input("number").
		Attr("min", "0").Attr("max", "150").Attr("step", "0.1").
		Attr("value", ageVal).
		Class(NameAgeDate.Class(PartAge).String()).
		OnInput(func(e dom.Event) {
			if date, ok := birthFromAge(e.TargetValue(), today()); ok {
				value.Set(date)
				if onInput != nil {
					onInput(date)
				}
			}
		})

	dateInput := Input("date").
		Attr("value", value.Get()).
		Class(NameAgeDate.Class(PartDate).String()).
		OnChange(func(e dom.Event) {
			val := e.TargetValue()
			value.Set(val)
			if age, ok := ageFromBirth(val, today()); ok {
				if ref, ok := ageInput.Ref(); ok {
					ref.SetValue(age)
				}
			} else {
				if ref, ok := ageInput.Ref(); ok {
					ref.SetValue("")
				}
			}
			if onInput != nil {
				onInput(val)
			}
		}).
		Bind(value)

	ageId := "age-" + a.GetID()
	ageInput.ID(ageId)

	return Div().Class(NameAgeDate.Root().String()).Child(
		Label().For(ageInput).Text(string(lang.Text("Age"))),
		ageInput,
		dateInput,
	)
}
