package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

type Environment map[string]EnvValue

// EnvValue helps to distinguish between empty files and files with the first empty line.
type EnvValue struct {
	Value      string
	NeedRemove bool
}

var (
	ErrBadDirPath          = errors.New("bad path to env dir")
	ErrNotDirectory        = errors.New("dir is not directory")
	ErrDirectoryUreadable  = errors.New("cant read directory")
	ErrEqualSignInFileName = errors.New("= in file name")
)

// ReadDir reads a specified directory and returns map of env variables.
// Variables represented as files where filename is name of variable, file first line is a value.
func ReadDir(dir string) (Environment, error) {
	if len(dir) == 0 {
		return nil, ErrBadDirPath
	}

	dirStat, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("cant get stat for env dir, err: %w", errors.Join(err, ErrBadDirPath))
	}

	if !dirStat.IsDir() {
		return nil, ErrNotDirectory
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, ErrDirectoryUreadable
	}

	envs := make(Environment)
	for _, entry := range entries {
		if !entry.Type().IsRegular() {
			continue
		}

		if strings.Contains(entry.Name(), "=") {
			return nil, fmt.Errorf("bad file name %s, err :%w", entry.Name(), ErrEqualSignInFileName)
		}

		rawValue, err := readline(dir + "/" + entry.Name())
		if err != nil {
			return nil, err
		}

		value := strings.TrimRight(string(bytes.ReplaceAll(rawValue, []byte{0x00}, []byte{'\n'})), "\t \n")
		envs[entry.Name()] = EnvValue{
			Value:      value,
			NeedRemove: len(rawValue) == 0,
		}
	}

	return envs, nil
}

func readline(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("cant read file, err: %w", err)
	}
	defer file.Close()

	br := bufio.NewReader(file)
	line, err := br.ReadBytes('\n')

	if err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("err while reading line, err: %w", err)
	}

	return line, nil
}
