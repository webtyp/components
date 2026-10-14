# ContentCard

Container with header, body, and footer sections. The header title uses `style.Heading`.

## Import
`"webtyp.com/components/contentcard"`

## Usage

```go
import (
    "webtyp.com/components/contentcard"
    . "webtyp.com/html"
)

c := &contentcard.ContentCard{
    Header: H3("Card Title"),
    Body:   P("Content goes here..."),
    Footer: Small("Last updated: today"),
}
```

## Properties
- `Header` (dom.Component): Content for header section.
- `Body` (dom.Component): Main content.
- `Footer` (dom.Component): Footer content.

---
[Back to Catalog](../docs/CATALOG.md)
