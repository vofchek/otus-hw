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
	tests := []struct {
		Case  string
		Level string
		LogFn func(logger *Logger, msg string)
	}{
		{
			"info method",
			"info",
			func(logger *Logger, msg string) { logger.Info(msg) },
		},
		{
			"error method",
			"error",
			func(logger *Logger, msg string) { logger.Error(msg) },
		},
		{
			"log method with info level",
			"info",
			func(logger *Logger, msg string) { logger.Log(t.Context(), infoLevel, msg) },
		},
		{
			"log method with error level",
			"error",
			func(logger *Logger, msg string) { logger.Log(t.Context(), errorLevel, msg) },
		},
	}

	for _, test := range tests {
		t.Run(test.Case, func(t *testing.T) {
			testDir := t.TempDir()
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
			assert.NoError(t, json.Unmarshal(outData, &log), "bad json: %s", outData)
			assert.Equal(t, test.Level, strings.ToLower(log.Level))
		})
	}

	t.Run("broken log level", func(t *testing.T) {
		testDir := t.TempDir()
		out := fmt.Sprintf("%s/calendar_log_%s.txt", testDir, "broken")
		_, err := New("broken", out)
		require.ErrorIs(t, err, ErrInvalidLogLevel)
	})
}

func TestFilterTestLevel(t *testing.T) {
	testDir := t.TempDir()

	t.Run("filter test level", func(t *testing.T) {
		testMsg := "this message must not be in log file"

		out := fmt.Sprintf("%s/calendar_log_%s.txt", testDir, "error")
		logg, err := New("error", out)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, logg.Close())
		})
		logg.Info(testMsg)
		outData, err := os.ReadFile(out)
		if err != nil {
			t.Fatal("could not open out file")
		}
		require.Empty(t, outData, "log data must be empty")
	})
}
