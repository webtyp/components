# AgeDate

At a clinic's front desk, people are typically asked their **age**, not their exact birth date. This component provides an input field that accepts both. Typing an age automatically computes and fills the birth date field (relative to today).

Only the resulting date (YYYY-MM-DD) is stored as the value, because an age changes every year.

Example:
```go
model.Definition{
    Name: "birthdate",
    Type: agedate.New(),
    Label: "Birthdate",
    OmitEmpty: true,
}
```
