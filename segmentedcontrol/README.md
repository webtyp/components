# SegmentedControl

A horizontal pill switch for 2 or more options.

## Usage

```go
comp := &segmentedcontrol.SegmentedControl{
	Options: []segmentedcontrol.Option{{Value: "A", Label: "A"}},
	Selected: "A",
	OnChange: func(val string) { println(val) },
}
```
