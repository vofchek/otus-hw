package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogger(t *testing.T) {
	testDir := t.TempDir()

	tests := []struct {
		Level string
		LogFn func(logger *Logger, msg string)
	}{
		{"info", func(logger *Logger, msg string) { logger.Info(msg) }},
		{"error", func(logger *Logger, msg string) { logger.Error(msg) }},
	}

	for _, test := range tests {
		t.Run(test.Level, func(t *testing.T) {
			out := fmt.Sprintf("%s/calendar_log_%s.txt", testDir, test.Level)

			l, err := New(test.Level, out)
			assert.NoError(t, err)
			t.Cleanup(func() {
				require.NoError(t, l.Close())
			})
			test.LogFn(l, test.Level)

			outData, err := os.ReadFile(out)
			if err != nil {
				t.Fatal("could not open out file")
			}

			var log struct {
				Level string
			}
			assert.NoError(t, json.Unmarshal(outData, &log))
			assert.Equal(t, test.Level, strings.ToLower(log.Level))
		})
	}
}
