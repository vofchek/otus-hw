package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	testKey   = "TEST_KEY"
	testValue = "test-value"
)

func TestReadDir(t *testing.T) {
	t.Run("bad directory path", func(t *testing.T) {
		dir := t.TempDir()
		in, err := os.Create(dir + "/empty_in.txt")
		if err != nil {
			t.Fatal(err)
		}
		defer in.Close()

		_, err = ReadDir(dir + "/empty_in.txt")
		require.ErrorIs(t, err, ErrNotDirectory)

		_, err = ReadDir("./____dir_not_exists_____")
		require.ErrorIs(t, err, ErrBadDirPath)
	})

	t.Run("make env from file names", func(t *testing.T) {
		dir := t.TempDir()

		in, err := os.Create(dir + "/" + testKey)
		if err != nil {
			t.Fatal(err)
		}
		in.WriteString(testValue)
		in.WriteString("\t\t\t\t\t              ")
		in.WriteString("\n")
		in.WriteString("------------------------")
		defer in.Close()

		env, err := ReadDir(dir)
		require.Nil(t, err)
		require.Contains(t, env, testKey)
		require.False(t, env[testKey].NeedRemove)
		require.Equal(t, testValue, env[testKey].Value)
	})

	t.Run("make env from file names", func(t *testing.T) {
		dir := t.TempDir()

		in, err := os.Create(dir + "/" + testKey)
		if err != nil {
			t.Fatal(err)
		}
		defer in.Close()

		in.WriteString(testValue)
		in.WriteString("\t\t\t\t\t              ")

		env, err := ReadDir(dir)
		require.Nil(t, err)
		require.Contains(t, env, testKey)
		require.False(t, env[testKey].NeedRemove)
		require.Equal(t, testValue, env[testKey].Value)
	})

	t.Run("empty env file - need to remove", func(t *testing.T) {
		testKey := "TEST_ENV_NEED_TO_REMOVE"

		dir := t.TempDir()

		in, err := os.Create(dir + "/" + testKey)
		if err != nil {
			t.Fatal(err)
		}
		defer in.Close()

		env, err := ReadDir(dir)
		require.Nil(t, err)
		require.Contains(t, env, testKey)
		require.True(t, env[testKey].NeedRemove)
	})

	t.Run("file name with =", func(t *testing.T) {
		dir := t.TempDir()

		in, err := os.Create(dir + "/TEST=")
		if err != nil {
			t.Fatal(err)
		}
		defer in.Close()

		_, err = ReadDir(dir)
		require.ErrorIs(t, err, ErrEqualSignInFileName)
	})

	t.Run("zero-byte in env value", func(t *testing.T) {
		dir := t.TempDir()
		testValue := "test" + "\x00" + "value"
		testValueExpected := "test" + "\n" + "value"

		in, err := os.Create(dir + "/" + testKey)
		if err != nil {
			t.Fatal(err)
		}
		defer in.Close()

		in.WriteString(testValue)

		env, err := ReadDir(dir)
		require.Nil(t, err)
		require.Contains(t, env, testKey)
		require.Equal(t, testValueExpected, env[testKey].Value)
	})
}
