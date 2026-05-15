package tui

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/manojshevate/mcp-setu/internal/bridge"
	"github.com/manojshevate/mcp-setu/internal/ollama"
)

// nonInteractivePrinter prints directly to stdout for non-TTY environments.
type nonInteractivePrinter struct {
	verbose bool
}

func (p *nonInteractivePrinter) PrintLLMProcessing(iteration int) {
	if !p.verbose {
		return
	}
	if iteration == 1 {
		fmt.Fprintln(os.Stderr, "Processing...")
	} else {
		fmt.Fprintf(os.Stderr, "Processing (iteration %d)...\n", iteration)
	}
}

func (p *nonInteractivePrinter) PrintWarning(msg string) {
	fmt.Fprintf(os.Stderr, "warning: %s\n", msg)
}

func (p *nonInteractivePrinter) PrintResponseStart() {}

func (p *nonInteractivePrinter) PrintResponseChunk(chunk string) {
	fmt.Print(chunk)
}

func (p *nonInteractivePrinter) PrintResponseEnd() {
	fmt.Println()
}

func (p *nonInteractivePrinter) PrintToolCall(name string, args map[string]any) {
	if !p.verbose {
		return
	}
	fmt.Fprintf(os.Stderr, "tool call: %s\n", name)
}

func (p *nonInteractivePrinter) PrintToolResult(name string, result string, truncated bool) {
	if !p.verbose {
		return
	}
	display := result
	if len(display) > 120 {
		display = display[:120] + "..."
	}
	fmt.Fprintf(os.Stderr, "tool result: %s\n", display)
}

// RunNonInteractive runs a simple line-based chat loop for non-TTY environments
// (e.g., piped input, CI, scripts).
func RunNonInteractive(
	ctx context.Context,
	br *bridge.Bridge,
	systemPrompt string,
	verbose bool,
) error {
	printer := &nonInteractivePrinter{verbose: verbose}
	br.SetPrinter(printer)

	history := []ollama.Message{{Role: "system", Content: systemPrompt}}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if line == "/quit" || line == "/exit" || line == "quit" || line == "exit" {
			break
		}

		history = append(history, ollama.Message{Role: "user", Content: line})
		content, err := br.ProcessMessage(ctx, history)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			history = history[:len(history)-1] // remove failed user message
			continue
		}
		history = append(history, ollama.Message{Role: "assistant", Content: content})

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("input error: %w", err)
	}
	return nil
}
