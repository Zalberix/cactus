package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	stdLog "log"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
)

const (
	EnvLocal = "local"
	EnvDev   = "dev"
	EnvProd  = "prod"

	maxFieldValueLen = 2048
	maxLineLen       = 8192
)

type formattedAttr struct {
	key   string
	value slog.Value
}

func SetupLogger(env string, source ...string) *slog.Logger {
	level := slog.LevelInfo
	if env == EnvLocal || env == EnvDev {
		level = slog.LevelDebug
	}

	name := ""
	if len(source) > 0 {
		name = source[0]
	}

	return slog.New(newTextHandler(os.Stdout, level, name, env == EnvLocal || env == EnvDev))
}

func WorkerSource(workerType string, workerID int32) string {
	if workerID <= 0 {
		return "worker/" + workerType
	}
	return "worker/" + workerType + "-" + strconv.FormatInt(int64(workerID), 10)
}

type textHandler struct {
	l        *stdLog.Logger
	level    slog.Level
	source   string
	colorize bool
	attrs    []formattedAttr
	groups   []string
}

func newTextHandler(out io.Writer, level slog.Level, source string, colorize bool) *textHandler {
	return &textHandler{
		l:        stdLog.New(out, "", 0),
		level:    level,
		source:   source,
		colorize: colorize,
	}
}

func (h *textHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *textHandler) Handle(_ context.Context, r slog.Record) error {
	fields := make([]string, 0, len(h.attrs)+r.NumAttrs())
	truncated := false

	for _, a := range h.attrs {
		addAttr(&fields, a.key, a.value, &truncated)
	}

	r.Attrs(func(a slog.Attr) bool {
		addSlogAttr(&fields, h.prefixedKey(a.Key), a.Value, &truncated)
		return true
	})

	parts := []string{
		r.Time.Format("[15:04:05.000]"),
		h.formatLevel(r.Level),
	}
	if h.source != "" {
		parts = append(parts, h.formatSource(h.source))
	}
	parts = append(parts, sanitizeText(r.Message))
	parts = append(parts, fields...)

	line := strings.Join(parts, " ")
	if truncated {
		line += " truncated=true"
	}
	if len(line) > maxLineLen {
		line = line[:maxLineLen] + "... truncated=true"
	}

	h.l.Println(line)
	return nil
}

func (h *textHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	next := h.clone()
	for _, a := range attrs {
		next.addInheritedAttr(h.prefixedKey(a.Key), a.Value)
	}
	return next
}

func (h *textHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	next := h.clone()
	next.groups = append(next.groups, name)
	return next
}

func (h *textHandler) clone() *textHandler {
	next := *h
	next.attrs = append([]formattedAttr(nil), h.attrs...)
	next.groups = append([]string(nil), h.groups...)
	return &next
}

func (h *textHandler) addInheritedAttr(key string, value slog.Value) {
	value = value.Resolve()
	if value.Kind() == slog.KindGroup {
		for _, a := range value.Group() {
			h.addInheritedAttr(joinKey(key, a.Key), a.Value)
		}
		return
	}
	h.attrs = append(h.attrs, formattedAttr{key: key, value: value})
}

func (h *textHandler) prefixedKey(key string) string {
	for i := len(h.groups) - 1; i >= 0; i-- {
		key = joinKey(h.groups[i], key)
	}
	return key
}

func (h *textHandler) formatLevel(level slog.Level) string {
	text := level.String()
	if !h.colorize {
		return text
	}

	switch {
	case level >= slog.LevelError:
		return color.New(color.FgRed, color.Bold).Sprint(text)
	case level >= slog.LevelWarn:
		return color.New(color.FgYellow, color.Bold).Sprint(text)
	case level >= slog.LevelInfo:
		return color.New(color.FgBlue, color.Bold).Sprint(text)
	default:
		return color.New(color.FgMagenta, color.Bold).Sprint(text)
	}
}

func (h *textHandler) formatSource(source string) string {
	text := "[" + source + "]"
	if !h.colorize {
		return text
	}
	return color.New(color.FgCyan, color.Bold).Sprint(text)
}

func addSlogAttr(fields *[]string, key string, value slog.Value, truncated *bool) {
	value = value.Resolve()
	if value.Kind() == slog.KindGroup {
		for _, a := range value.Group() {
			addSlogAttr(fields, joinKey(key, a.Key), a.Value, truncated)
		}
		return
	}
	addAttr(fields, key, value, truncated)
}

func addAttr(fields *[]string, key string, value slog.Value, truncated *bool) {
	if key == "" {
		return
	}
	formatted := formatValue(value, truncated)
	*fields = append(*fields, key+"="+formatted)
}

func formatValue(value slog.Value, truncated *bool) string {
	value = value.Resolve()

	var raw string
	switch value.Kind() {
	case slog.KindString:
		raw = value.String()
	case slog.KindBool:
		raw = strconv.FormatBool(value.Bool())
	case slog.KindInt64:
		raw = strconv.FormatInt(value.Int64(), 10)
	case slog.KindUint64:
		raw = strconv.FormatUint(value.Uint64(), 10)
	case slog.KindFloat64:
		raw = strconv.FormatFloat(value.Float64(), 'g', -1, 64)
	case slog.KindDuration:
		raw = value.Duration().String()
	case slog.KindTime:
		raw = value.Time().Format(time.RFC3339Nano)
	case slog.KindAny:
		raw = formatAny(value.Any())
	default:
		raw = value.String()
	}

	if len(raw) > maxFieldValueLen {
		raw = raw[:maxFieldValueLen] + "..."
		*truncated = true
	}

	return quoteIfNeeded(raw)
}

func formatAny(v any) string {
	if v == nil {
		return "<nil>"
	}
	if data, err := json.Marshal(v); err == nil {
		return string(data)
	}
	return fmt.Sprint(v)
}

func quoteIfNeeded(s string) string {
	if s == "" {
		return `""`
	}
	if strings.ContainsAny(s, " \t\r\n\"=") {
		return strconv.Quote(s)
	}
	return sanitizeText(s)
}

func sanitizeText(s string) string {
	s = strings.ReplaceAll(s, "\r", `\r`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	return s
}

func joinKey(prefix, key string) string {
	if prefix == "" {
		return key
	}
	if key == "" {
		return prefix
	}
	return prefix + "." + key
}
