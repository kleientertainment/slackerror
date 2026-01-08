package slack

import (
	"context"
	"log/slog"
	"maps"
)

type SlogHandler struct {
	channel *SlackChannel
	attrs   map[string]string
}

var logLevelColours = map[slog.Level]Colour{
	slog.LevelDebug: Good,
	slog.LevelInfo:  Good,
	slog.LevelWarn:  Warning,
	slog.LevelError: Danger,
}

func NewSlogHandler(channel *SlackChannel, baseFields map[string]string) *SlogHandler {
	return &SlogHandler{channel: channel, attrs: baseFields}
}

func (h *SlogHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= slog.LevelDebug // Send all levels to Slack
}

func (h *SlogHandler) Handle(_ context.Context, r slog.Record) error {
	fields := maps.Clone(h.attrs)
	r.Attrs(func(attr slog.Attr) bool {
		fields[attr.Key] = attr.Value.String()
		return true
	})
	return h.channel.SendMessage(r.Message, levelToColour(r.Level), fields, nil)
}

func (h *SlogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := maps.Clone(h.attrs)
	for _, attr := range attrs {
		newAttrs[attr.Key] = attr.Value.String()
	}
	return &SlogHandler{channel: h.channel, attrs: newAttrs}
}

func (h *SlogHandler) WithGroup(string) slog.Handler {
	// Not implemented
	return h
}

func levelToColour(level slog.Level) Colour {
	if colour, ok := logLevelColours[level]; ok {
		return colour
	}
	// Good colour as the default for any other levels
	return Good
}
