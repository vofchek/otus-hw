package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/cheggaaa/pb"
)

var (
	ErrUnsupportedFile            = errors.New("unsupported file")
	ErrOffsetExceedsFileSize      = errors.New("offset exceeds file size")
	ErrOffsetMustBeZeroOrPositive = errors.New("offset must be >= 0")
	ErrPathMustNotBeEmpty         = errors.New("path must not be empty")
	ErrSamePath                   = errors.New("from path and to path must not be equal")

	// размер порции скопированных данных.
	copyBatchSize int64 = 1000
	fullCopy      int64
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

	// подсчет total для полоски прогресса.
	copySize := int64(0)
	switch limit {
	case fullCopy:
		copySize = fromStat.Size() - offset
	default:
		copySize = min(fromStat.Size()-offset, limit)
	}

	bar := pb.New(int(copySize))

	// копирование файла кусками, пока не скопируем запрошенный limit.
	copied := int64(0)
	for copied < copySize {
		limitToCopy := min(copyBatchSize, copySize)

		n, err := io.CopyN(to, from, limitToCopy)
		copied += n
		bar.Add64(n)
		if err != nil && !errors.Is(err, io.EOF) {
			bar.FinishPrint("copy error!")
			return err
		}
	}
	bar.FinishPrint("copy done!")

	return nil
}
