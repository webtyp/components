# CellGrid

An interactive, fluid 2D coordinate grid canvas.

## Usage

```go
comp := &cellgrid.CellGrid{
	Cols: 5, Rows: 5,
	OnCellClick: func(r, c int) { println(r, c) },
}
```
