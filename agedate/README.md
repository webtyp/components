# AgeDate

At a clinic's front desk, people are typically asked their **age**, not their exact birth date. This input accepts both. Typing an age (whole years, **1 or more**) fills the birth date (today minus that many years) and, when the model names a field in `AlsoSetsField`, marks it as estimated; typing the exact date marks it as not estimated.

Under one year there is no age: the exact birth date is required, and the age box says so when it loses focus.

Only the date (YYYY-MM-DD) is stored, because an age changes every year.

```go
{Name: "birthdate", Type: agedate.New(), Label: "Birthdate", OmitEmpty: true, AlsoSetsField: "birthdate_estimated"},
{Name: "birthdate_estimated", Type: input.Checkbox(), Label: "Estimated birth date"},
```
