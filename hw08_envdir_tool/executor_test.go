package main

import (
	"bufio"
	"bytes"
	"os"
	"strings"
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

	t.Run("test command return error code", func(t *testing.T) {
		env := Environment{}
		code := RunCmd([]string{"ls", "/catalog_not_exists"}, env)

		require.Equal(t, 2, code)

		code = RunCmd([]string{"bash", "-c", "exit 67"}, env)

		require.Equal(t, 67, code)
	})

	t.Run("test stdin, stdout, stderr", func(t *testing.T) {
		testValue := "test\n"

		ecIn = strings.NewReader(testValue)
		defer func() { ecIn = os.Stdin }()

		buf := bytes.NewBuffer(make([]byte, 0, 5))
		ecOut = bufio.NewWriter(buf)
		defer func() { ecOut = os.Stdout }()

		env := Environment{}
		code := RunCmd([]string{"cat"}, env)

		require.Equal(t, codeDone, code)
		require.Equal(t, testValue, buf.String())
	})
}
