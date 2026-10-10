package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunCmd(t *testing.T) {
	t.Run("empty command", func(t *testing.T) {
		env := Environment{
			"TEST": {
				Value:      "TEST",
				NeedRemove: false,
			},
		}
		code := RunCmd([]string{}, env)

		require.Equal(t, codeErrEmptyCmd, code)
	})

	t.Run("command executed normally", func(t *testing.T) {
		env := Environment{
			"TEST": {
				Value:      "TEST",
				NeedRemove: false,
			},
		}
		code := RunCmd([]string{"echo", "1"}, env)

		require.Equal(t, codeDone, code)
	})

	t.Run("command not executed", func(t *testing.T) {
		env := Environment{}
		code := RunCmd([]string{"unknown_command"}, env)

		require.Equal(t, codeErrStartErr, code)
	})
}
