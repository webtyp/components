# InboxList

`InboxList` (`widget.Listbox`) renders a conversation list with title, message preview, timestamp, and unread badge.

## Usage

```go
import "webtyp.com/components/inboxlist"

list := &inboxlist.InboxList{
    Empty: "No conversations",
    OnSelect: func(id string) {
        fmt.Println("Selected conversation:", id)
    },
}

list.SetRows([]inboxlist.Row{
    {ID: "c1", Title: "Alice", Preview: "See you tomorrow!", Time: "14:32", Unread: 2},
    {ID: "c2", Title: "Bob", Preview: "Thanks!", Time: "11:05", Unread: 0},
})
```

## Fields

- `Selected *SignalString`: Optional signal holding the currently selected row ID.
- `OnSelect func(id string)`: Callback triggered when a row is clicked.
- `Empty string`: Text shown when there are no rows in the list.
