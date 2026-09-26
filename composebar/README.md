# ComposeBar

`ComposeBar` (`widget.Form`) renders a message input field with a send button and Enter-to-send keyboard handling.

## Usage

```go
import "webtyp.com/components/composebar"

bar := &composebar.ComposeBar{
    Placeholder: "Escribe un mensaje...",
    SendLabel:   "Enviar",
    MaxLength:   2000,
    OnSend: func(body string) {
        fmt.Println("Sending:", body)
    },
}
```

## Fields

- `Placeholder string`: Text shown inside the input when empty.
- `SendLabel string`: Label text for the send button.
- `MaxLength int`: Maximum character length allowed on input.
- `OnSend func(body string)`: Callback invoked with trimmed message text on submit (**required**, panics if nil).
- `Disabled *SignalBool`: Optional signal disabling both the input and send button.
