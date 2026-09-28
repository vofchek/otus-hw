package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
)

var (
	ErrUnsupportedFile            = errors.New("unsupported file")
	ErrOffsetExceedsFileSize      = errors.New("offset exceeds file size")
	ErrOffsetMustBeZeroOrPositive = errors.New("offset must be >= 0")
	ErrPathMustNotBeEmpty         = errors.New("path must not be empty")
	ErrSamePath                   = errors.New("from path and to path must not be equal")
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
	defer func() {
		err := from.Close()
		if err != nil {
			log.Printf("error while closing from: %v", err)
		}
	}()

	fromStat, err := from.Stat()
	if err != nil {
		return fmt.Errorf("failed get state for fromPath: %w", err)
	}

	if !fromStat.Mode().IsRegular() {
		return ErrUnsupportedFile
	}

	if fromStat.Size() < offset {
		return ErrOffsetExceedsFileSize
	}

	toStat, err := os.Stat(toPath)
	if err == nil && os.SameFile(toStat, fromStat) {
		return ErrSamePath
	}

	to, err := os.Create(toPath)
	if err != nil {
		return fmt.Errorf("create toPath err: %w", err)
	}
	defer func() {
		err := to.Close()
		if err != nil {
			log.Printf("error while closing to: %v", err)
		}
	}()

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
