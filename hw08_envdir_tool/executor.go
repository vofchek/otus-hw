package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	codeDone = iota
	codeErrEmptyCmd
	codeErrStartErr
)

// RunCmd runs a command + arguments (cmd) with environment variables from env.
func RunCmd(cmd []string, env Environment) (returnCode int) {
	if len(cmd) == 0 {
		returnCode = codeErrEmptyCmd
		return
	}

	returnCode = codeDone

	safePath := filepath.Clean(cmd[0])

	ec := exec.Command(safePath, cmd[1:]...)
	ec.Stdout = os.Stdout
	ec.Stderr = os.Stderr

	commandEnv := make([]string, 0, 5)

	for _, envLine := range os.Environ() {
		parts := strings.Split(envLine, "=")
		if len(parts) == 0 {
			continue
		}

		_, keyInEnv := env[parts[0]]
		if !keyInEnv || !env[parts[0]].NeedRemove {
			commandEnv = append(commandEnv, envLine)
		}
	}

	for eK, eV := range env {
		if !eV.NeedRemove {
			commandEnv = append(commandEnv, fmt.Sprintf("%s=%s", eK, eV.Value))
		}
	}

	ec.Env = commandEnv

	err := ec.Run()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			returnCode = exitErr.ExitCode()
		} else {
			returnCode = codeErrStartErr
		}
	}

	return returnCode
}
