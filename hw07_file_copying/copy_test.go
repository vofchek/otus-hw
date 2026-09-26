package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var inputFile = "./testdata/input.txt"

func TestCopy(t *testing.T) {
	t.Run("negative offset", func(t *testing.T) {
		err := Copy(inputFile, "/dev/null", -1, 1024)

		require.ErrorIs(t, err, ErrOffsetMustBeZeroOrPositive)
	})

	t.Run("offset more than file size", func(t *testing.T) {
		err := Copy(inputFile, "/dev/null", 100_000_000_000, 1024)

		require.ErrorIs(t, err, ErrOffsetExceedsFileSize)
	})

	t.Run("filesize unavailable", func(t *testing.T) {
		err := Copy("/dev/urandom", "/dev/null", 100_000_000_000, 1024)

		require.ErrorIs(t, err, ErrUnsupportedFile)
	})

	t.Run("empty frompath", func(t *testing.T) {
		err := Copy("", "/dev/null", 0, 1024)

		require.ErrorIs(t, err, ErrPathMustNotBeEmpty)
	})

	t.Run("empty topath", func(t *testing.T) {
		err := Copy(inputFile, "", 0, 1024)

		require.ErrorIs(t, err, ErrPathMustNotBeEmpty)
	})

	inStat, err := os.Stat(inputFile)
	require.Nil(t, err)

	testParams := []struct {
		Title        string
		Offset       int64
		Limit        int64
		ExpectedSize int64
	}{
		{"full copy", 0, 0, inStat.Size()},
		{"copy 10", 0, 10, 10},
		{"copy 1000", 0, 1000, 1000},
		{"copy 1000 from 100", 100, 1000, 1000},
		{"copy 1000 from 6000 offset", 6000, 1000, inStat.Size() - 6000},
	}

	for _, params := range testParams {
		t.Run(params.Title, func(t *testing.T) {
			out := fmt.Sprintf("./testdata/out_offset%d_limit%d.txt", params.Offset, params.Limit)

			copyErr := Copy(inputFile, out, params.Offset, params.Limit)

			outStat, err := os.Stat(out)
			assert.Nil(t, err, "cant open out file")

			assert.Nil(t, copyErr, "copy error")
			assert.Equal(t, params.ExpectedSize, outStat.Size(), "wrong out size")
		})
	}
}
