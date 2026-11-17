package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"go.opentelemetry.io/otel/trace"
)

const (
	// TraceIDKey is the key under which [From] adds the trace id of the active span, when a sink is set.
	TraceIDKey = "trace_id"
	// SpanIDKey is the key under which [From] adds the span id of the active span, when a sink is set.
	SpanIDKey = "span_id"
)

// Attr is a single key / value pair of a log [Entry]. Values are decoded from JSON,
// so numbers arrive as [json.Number].
type Attr struct {
	Key   string
	Value any
}

// Entry is a single log line, as handed to a [Sink].
type Entry struct {
	Time time.Time
	// Level is [log.InfoLevel] for lines without a level (log.Print).
	Level   log.Level
	Prefix  string
	Message string
	Caller  string
	Attrs   []Attr
}

// Sink receives every log line written through the default logger (and loggers derived from it).
// It is called concurrently and must not log through the default logger itself.
type Sink func(entry *Entry)

var (
	bridgeMu sync.RWMutex
	bridged  bool
)

// SetSink routes the default logger through sink, while still printing every line to the console.
//
// The default logger is switched to JSON so lines can be parsed back into structured entries; the console
// output keeps the configured format. Loggers derived from the default logger before this call are not routed.
func SetSink(sink Sink) {
	bridgeMu.Lock()
	defer bridgeMu.Unlock()

	if console == nil {
		console = newConsole(consoleOut)
	}

	logger := log.Default()
	logger.SetOutput(&bridgeWriter{console: console, out: consoleOut, sink: sink})
	logger.SetFormatter(log.JSONFormatter)
	logger.SetTimeFormat(time.RFC3339Nano)
	logger.SetReportTimestamp(true)
	bridged = true
}

// Console returns a logger that only prints to the console, never to the sink.
// Use it for errors of the sink itself, to avoid feeding them back into it.
func Console() *log.Logger {
	bridgeMu.RLock()
	defer bridgeMu.RUnlock()

	if !bridged {
		return log.Default()
	}

	return console.With() // a copy, so callers can't reconfigure the bridge's console
}

// withTrace adds the ids of the span in ctx to the logger, so the sink can correlate the line with its trace.
func withTrace(ctx context.Context, logger *log.Logger) *log.Logger {
	if ctx == nil || !isBridged() {
		return logger
	}

	sc := trace.SpanContextFromContext(ctx)
	if !sc.IsValid() {
		return logger
	}

	return logger.With(TraceIDKey, sc.TraceID().String(), SpanIDKey, sc.SpanID().String())
}

func isBridged() bool {
	bridgeMu.RLock()
	defer bridgeMu.RUnlock()

	return bridged
}

// bridgeWriter receives the JSON lines of the default logger, hands them to the sink and re-renders them on the console.
type bridgeWriter struct {
	console *log.Logger
	out     io.Writer
	sink    Sink
}

func (w *bridgeWriter) Write(p []byte) (int, error) {
	for line := range bytes.Lines(p) {
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		entry, hasLevel, err := parseEntry(line)
		if err != nil {
			// Not something we can parse, print it as-is rather than losing it
			if _, werr := w.out.Write(append(line, '\n')); werr != nil {
				return 0, werr
			}

			continue
		}

		w.sink(entry)
		w.print(entry, hasLevel)
	}

	return len(p), nil
}

// print renders the entry on the console logger, in the same order the default logger would have.
func (w *bridgeWriter) print(entry *Entry, hasLevel bool) {
	kvs := make([]any, 0, 10+len(entry.Attrs)*2)
	kvs = append(kvs, log.TimestampKey, entry.Time)
	if hasLevel {
		kvs = append(kvs, log.LevelKey, entry.Level)
	}
	if entry.Caller != "" {
		kvs = append(kvs, log.CallerKey, entry.Caller)
	}
	if entry.Prefix != "" {
		kvs = append(kvs, log.PrefixKey, entry.Prefix)
	}
	if entry.Message != "" {
		kvs = append(kvs, log.MessageKey, entry.Message)
	}
	for _, attr := range entry.Attrs {
		if attr.Key == TraceIDKey || attr.Key == SpanIDKey {
			continue // only useful to the sink, keep the console readable
		}
		kvs = append(kvs, attr.Key, attr.Value)
	}

	// Print has no level of its own, so the keyvals above fully control the output
	w.console.Print(nil, kvs...)
}

func parseEntry(line []byte) (*Entry, bool, error) {
	dec := json.NewDecoder(bytes.NewReader(line))
	dec.UseNumber()

	if tok, err := dec.Token(); err != nil || tok != json.Delim('{') {
		return nil, false, errors.New("log line is not a JSON object")
	}

	entry := &Entry{Level: log.InfoLevel}
	hasLevel := false
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return nil, false, fmt.Errorf("failed to read log key: %w", err)
		}
		key, ok := tok.(string)
		if !ok {
			return nil, false, fmt.Errorf("unexpected log key: %v", tok)
		}

		var value any
		if err := dec.Decode(&value); err != nil {
			return nil, false, fmt.Errorf("failed to read log value of %s: %w", key, err)
		}

		switch key {
		case log.TimestampKey:
			if s, ok := value.(string); ok {
				if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
					entry.Time = t
				}
			}
		case log.LevelKey:
			if s, ok := value.(string); ok {
				if level, err := log.ParseLevel(s); err == nil {
					entry.Level = level
					hasLevel = true
				}
			}
		case log.CallerKey:
			entry.Caller, _ = value.(string)
		case log.PrefixKey:
			entry.Prefix, _ = value.(string)
		case log.MessageKey:
			entry.Message, _ = value.(string)
		default:
			entry.Attrs = append(entry.Attrs, Attr{Key: key, Value: value})
		}
	}

	if entry.Time.IsZero() {
		entry.Time = time.Now()
	}

	return entry, hasLevel, nil
}
