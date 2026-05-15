package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Logger struct {
	Enabled   bool // Exported for testing
	sessionID string
	logPath   string
	file      *os.File
	mu        sync.Mutex
}

// New creates a new logger with a random session ID.
// If enabled is false, logging is disabled but the logger is still functional (no-op).
func New(enabled bool) (*Logger, error) {
	if !enabled {
		return &Logger{Enabled: false}, nil
	}

	sessionID := generateSessionID()
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("could not determine home directory: %w", err)
	}

	logDir := filepath.Join(homeDir, ".mcp-setu")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("could not create log directory: %w", err)
	}

	logPath := filepath.Join(logDir, sessionID+".log")
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return nil, fmt.Errorf("could not open log file: %w", err)
	}

	return &Logger{
		Enabled:   true,
		sessionID: sessionID,
		logPath:   logPath,
		file:      file,
	}, nil
}

// LogUserMessage logs a user message with timestamp.
func (l *Logger) LogUserMessage(content string) {
	if !l.Enabled {
		return
	}
	l.write("USER_MESSAGE", content)
}

// LogLLMResponse logs the LLM response with timestamp.
func (l *Logger) LogLLMResponse(content string) {
	if !l.Enabled {
		return
	}
	l.write("LLM_RESPONSE", content)
}

// LogToolCall logs a tool call request with arguments.
func (l *Logger) LogToolCall(name string, args map[string]any) {
	if !l.Enabled {
		return
	}
	argsStr := fmt.Sprintf("%v", args)
	l.write("TOOL_CALL_REQUEST", fmt.Sprintf("tool=%s args=%s", name, argsStr))
}

// LogToolResult logs a tool execution result.
func (l *Logger) LogToolResult(name string, result string, success bool) {
	if !l.Enabled {
		return
	}
	status := "success"
	if !success {
		status = "failed"
	}
	l.write("TOOL_CALL_RESULT", fmt.Sprintf("tool=%s status=%s result=%s", name, status, result))
}

// LogError logs an error message.
func (l *Logger) LogError(msg string) {
	if !l.Enabled {
		return
	}
	l.write("ERROR", msg)
}

// LogInfo logs an informational message.
func (l *Logger) LogInfo(msg string) {
	if !l.Enabled {
		return
	}
	l.write("INFO", msg)
}

// SessionID returns the session ID.
func (l *Logger) SessionID() string {
	return l.sessionID
}

// Close closes the log file.
func (l *Logger) Close() error {
	if !l.Enabled || l.file == nil {
		return nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.file.Close()
}

// write writes a log entry with a timestamp.
func (l *Logger) write(level string, msg string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	logEntry := fmt.Sprintf("[%s] %s: %s\n", timestamp, level, msg)

	if _, err := l.file.WriteString(logEntry); err != nil {
		// Silent fail — don't let logging errors break the app
		_ = err
	}
}

// generateSessionID generates a random session ID (timestamp + random suffix).
func generateSessionID() string {
	return fmt.Sprintf("session-%d", time.Now().UnixNano())
}
