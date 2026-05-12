package tui

import (
	"fmt"
)

// EventKind classifies a TUI output event.
type EventKind int

const (
	EventLog EventKind = iota // append a line to the output area
	EventRespStart
	EventRespChunk
	EventRespEnd
	EventWarning
	EventError
)

// Event is a single output emission from the bridge or commands.
type Event struct {
	Kind EventKind
	Text string
}

// Printer is the TUI's event-emitting printer. It satisfies bridge.Printer.
// Every method enqueues an Event on the channel and returns immediately.
// It does NOT write to stdout/stderr — all output is rendered by the TUI.
type Printer struct {
	ch chan<- Event
}

// NewPrinter creates a TUI printer that emits events to the given channel.
func NewPrinter(ch chan<- Event) *Printer {
	return &Printer{ch: ch}
}

func (p *Printer) emit(ev Event) {
	// Non-blocking send: if the channel is full, drop the event rather than
	// stall the bridge. The channel is sized generously in the session model.
	select {
	case p.ch <- ev:
	default:
	}
}

func (p *Printer) PrintLLMProcessing(iteration int) {
	if iteration == 1 {
		p.emit(Event{Kind: EventLog, Text: "💭 Processing..."})
	} else {
		p.emit(Event{Kind: EventLog, Text: fmt.Sprintf("💭 Processing (iteration %d)...", iteration)})
	}
}

func (p *Printer) PrintWarning(msg string) {
	p.emit(Event{Kind: EventWarning, Text: msg})
}

func (p *Printer) PrintResponseStart() {
	p.emit(Event{Kind: EventRespStart})
}

func (p *Printer) PrintResponseChunk(chunk string) {
	p.emit(Event{Kind: EventRespChunk, Text: chunk})
}

func (p *Printer) PrintResponseEnd() {
	p.emit(Event{Kind: EventRespEnd})
}

func (p *Printer) PrintToolCall(name string, args map[string]any) {
	p.emit(Event{Kind: EventLog, Text: fmt.Sprintf("⚙ %s", name)})
}

func (p *Printer) PrintToolResult(name string, result string, truncated bool) {
	display := result
	if len(display) > 120 {
		display = display[:120] + "…"
	}
	p.emit(Event{Kind: EventLog, Text: fmt.Sprintf("↳ %s", display)})
}
