package logger

import (
	"context"
	"encoding/json"
	"io"
	stdLog "log"
	"log/slog"
	"os"

	"github.com/fatih/color"
)

const (
	EnvLocal = "local"
	EnvDev   = "dev"
	EnvProd  = "prod"
)

func SetupLogger(env string) *slog.Logger {
	switch env {
	case EnvLocal:
		return slog.New(newPrettyHandler(os.Stdout, slog.LevelDebug))
	case EnvDev:
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	default:
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
}

type prettyHandler struct {
	slog.Handler
	l     *stdLog.Logger
	attrs []slog.Attr
}

func newPrettyHandler(out io.Writer, level slog.Level) *prettyHandler {
	return &prettyHandler{
		Handler: slog.NewJSONHandler(out, &slog.HandlerOptions{Level: level}),
		l:       stdLog.New(out, "", 0),
	}
}

func (h *prettyHandler) Handle(_ context.Context, r slog.Record) error {
	level := r.Level.String() + ":"
	switch r.Level {
	case slog.LevelDebug:
		level = color.MagentaString(level)
	case slog.LevelInfo:
		level = color.BlueString(level)
	case slog.LevelWarn:
		level = color.YellowString(level)
	case slog.LevelError:
		level = color.RedString(level)
	}

	fields := make(map[string]interface{}, r.NumAttrs())
	r.Attrs(func(a slog.Attr) bool {
		fields[a.Key] = a.Value.Any()
		return true
	})
	for _, a := range h.attrs {
		fields[a.Key] = a.Value.Any()
	}

	var b []byte
	if len(fields) > 0 {
		var err error
		b, err = json.MarshalIndent(fields, "", "  ")
		if err != nil {
			return err
		}
	}

	h.l.Println(
		r.Time.Format("[15:05:05.000]"),
		level,
		color.CyanString(r.Message),
		color.WhiteString(string(b)),
	)
	return nil
}

func (h *prettyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &prettyHandler{Handler: h.Handler, l: h.l, attrs: attrs}
}

func (h *prettyHandler) WithGroup(name string) slog.Handler {
	return &prettyHandler{Handler: h.Handler.WithGroup(name), l: h.l}
}
