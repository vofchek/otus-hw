package logger

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogger(t *testing.T) {
	buf := &bytes.Buffer{}

	tests := []struct {
		Level string
		LogFn func(logger *Logger, msg string)
	}{
		{"info", func(logger *Logger, msg string) { logger.Info(msg) }},
		{"error", func(logger *Logger, msg string) { logger.Error(msg) }},
	}

	for _, test := range tests {
		t.Run(test.Level, func(t *testing.T) {
			l := New(test.Level, buf)
			test.LogFn(l, test.Level)

			var log struct {
				Level string
			}
			assert.NoError(t, json.Unmarshal(buf.Bytes(), &log))
			assert.Equal(t, test.Level, strings.ToLower(log.Level))

			buf.Reset()
		})
	}
}
