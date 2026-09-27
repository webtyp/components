# BubbleThread

`BubbleThread` (`widget.Region`, `role="log"`, `aria-live="polite"`) renders a message thread as chat bubbles with distinct styling for own and received messages.

## Usage

```go
import "webtyp.com/components/bubblethread"

thread := &bubblethread.BubbleThread{
    ReadLabel: "Leído",
    Empty:     "No hay mensajes",
}

thread.SetBubbles([]bubblethread.Bubble{
    {ID: "1", Author: "Alice", Body: "Hola!", Time: "10:00", Mine: false},
    {ID: "2", Body: "Hola Alice, ¿cómo estás?", Time: "10:01", Mine: true, Read: true},
})

thread.Append(bubblethread.Bubble{
    ID: "3", Author: "Alice", Body: "Todo bien, ¿y tú?", Time: "10:02", Mine: false,
})
```

## Fields & Methods

- `ReadLabel string`: Label shown under read sent messages (e.g., "Leído").
- `Empty string`: Text displayed when there are no messages in the thread.
- `SetBubbles(b []Bubble)`: Replaces all bubbles in the thread.
- `Append(b ...Bubble)`: Appends new bubbles to the end, skipping duplicate IDs.
- `MarkRead(ids ...string)`: Marks specified sent message IDs as read.
- `Bubbles() []Bubble`: Returns current bubble records.
