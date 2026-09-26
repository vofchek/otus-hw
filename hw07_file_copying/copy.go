package main

import (
	"errors"
	"fmt"
	"io"
	"os"
)

var (
	ErrUnsupportedFile            = errors.New("unsupported file")
	ErrOffsetExceedsFileSize      = errors.New("offset exceeds file size")
	ErrOffsetMustBeZeroOrPositive = errors.New("offset must be >= 0")
	ErrPathMustNotBeEmpty         = errors.New("path must not be empty")
)

func Copy(fromPath, toPath string, offset, limit int64) error {
	if offset < 0 {
		return ErrOffsetMustBeZeroOrPositive
	}

	if len(fromPath) == 0 {
		return ErrPathMustNotBeEmpty
	}

	if len(toPath) == 0 {
		return ErrPathMustNotBeEmpty
	}

	from, err := os.Open(fromPath)
	if err != nil {
		return err
	}
	defer closeOrPanic(from)

	fromStat, err := from.Stat()
	if err != nil {
		return fmt.Errorf("failed get state for fromPath: %w", err)
	}

	if fromStat.Size() == 0 {
		return fmt.Errorf("file is zero sized: %w", ErrUnsupportedFile)
	}

	if fromStat.Size() < offset {
		return ErrOffsetExceedsFileSize
	}

	to, err := os.Create(toPath)
	if err != nil {
		return fmt.Errorf("create toPath err: %w", err)
	}
	defer closeOrPanic(to)

	if offset != 0 {
		_, err := from.Seek(offset, io.SeekStart)
		if err != nil {
			return fmt.Errorf("failed seek: %w", err)
		}
	}

	if limit == 0 {
		_, err = io.Copy(to, from)
		return err
	}

	_, err = io.CopyN(to, from, limit)
	if !errors.Is(err, io.EOF) {
		return err
	}

	return nil
}

func closeOrPanic(f io.Closer) {
	err := f.Close()
	if err != nil {
		panic(err)
	}
}
