# PresenceList

`PresenceList` (`widget.Listbox`) renders a list of contacts with an online/offline status dot and visually hidden screen-reader status text.

## Usage

```go
import "webtyp.com/components/presencelist"

list := &presencelist.PresenceList{
    OnlineLabel:  "En línea",
    OfflineLabel: "Desconectado",
    Empty:        "No hay contactos",
    OnSelect: func(id string) {
        fmt.Println("Selected contact:", id)
    },
}

list.SetPeople([]presencelist.Person{
    {ID: "p1", Label: "Alice", Online: true},
    {ID: "p2", Label: "Bob", Online: false},
})
```

## Fields & Methods

- `OnlineLabel string`: Screen-reader text for online status (e.g. "En línea").
- `OfflineLabel string`: Screen-reader text for offline status (e.g. "Desconectado").
- `Empty string`: Text shown when there are no contacts.
- `OnSelect func(id string)`: Callback fired when a contact row is clicked.
- `SetPeople(p []Person)`: Sets the list of contacts (sorted online first, then alphabetically by label).
- `People() []Person`: Returns the current contact records.
