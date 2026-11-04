package agedate

import (
	"math"
	"strconv"
	"strings"
	"webtyp.com/dom"
	"webtyp.com/fmt"

	"webtyp.com/input"
	"webtyp.com/time"
	"webtyp.com/widget"
	"webtyp.com/lang"
	. "webtyp.com/html"
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

// birthFromAge: today minus years and months. ageText uses '.' or ',' as decimal mark. "" or "0" → "", false.
func birthFromAge(ageText string, today time.Unix) (date string, ok bool) {
	if ageText == "" || ageText == "0" || ageText == "0.0" || ageText == "0,0" {
		return "", false
	}
	ageText = strings.ReplaceAll(ageText, ",", ".")
	val, err := strconv.ParseFloat(ageText, 64)
	if err != nil || val <= 0 {
		return "", false
	}

	years := int(math.Floor(val))
	fraction := val - math.Floor(val)
	months := int(math.Floor(fraction * 12))

	ty, tm, td, _, _, _ := time.CivilDate(today)

	newMonth := tm - months
	newYear := ty - years
	for newMonth <= 0 {
		newMonth += 12
		newYear--
	}

	tUnix := time.Date(newYear, newMonth, td, 0, 0, 0)

	cy, cm, cd, _, _, _ := time.CivilDate(tUnix)
	dateStr := fmt.Sprintf("%04d-%02d-%02d", cy, cm, cd)
	return dateStr, true
}

// ageFromBirth: whole years, plus the year fraction of full months, rounded to 1 decimal ("40.3").
func ageFromBirth(date string, today time.Unix) (age string, ok bool) {
	if date == "" {
		return "", false
	}

	t, err := time.ParseDate(date)
	if err != nil {
		return "", false
	}

	ty, tm, td, _, _, _ := time.CivilDate(today)
	by, bm, bd, _, _, _ := time.CivilDate(t)

	years := ty - by
	months := tm - bm
	days := td - bd

	if months < 0 || (months == 0 && days < 0) {
		years--
		months += 12
	}
	if days < 0 {
		months--
	}
	if months < 0 {
		months += 12
	}

	fraction := float64(months) / 12.0
	ageVal := float64(years) + fraction

	ageStr := fmt.Sprintf("%.1f", ageVal)
	return ageStr, true
}

func (a *AgeDate) RenderInput(value *dom.SignalString, onInput func(string)) *dom.Element {
	ageVal := ""
	if v := value.Get(); v != "" {
		if age, ok := ageFromBirth(v, time.NowUnix()); ok {
			ageVal = age
		}
	}

	ageInput := Input("number").
		Attr("min", "0").Attr("max", "150").Attr("step", "0.1").
		Attr("value", ageVal).
		Class(NameAgeDate.Class(PartAge).String()).
		OnInput(func(e dom.Event) {
			if date, ok := birthFromAge(e.TargetValue(), time.NowUnix()); ok {
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
			if age, ok := ageFromBirth(val, time.NowUnix()); ok {
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
