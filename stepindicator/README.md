# StepIndicator

A numbered wizard/stepper control.

## Usage

```go
comp := &stepindicator.StepIndicator{
	Steps: []stepindicator.Step{
		{Key: "1", Label: "One"},
		{Key: "2", Label: "Two"},
	},
	Active: 0,
	OnChange: func(idx int) { println("Active step:", idx) },
}
```
