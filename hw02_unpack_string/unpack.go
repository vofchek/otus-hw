package hw02unpackstring

import (
	"errors"
	"strconv"
	"strings"
)

var ErrInvalidString = errors.New("invalid string")

func Unpack(packed string) (string, error) {
	var unpacked = strings.Builder{}
	var prevSymbol string

	for _, r := range packed {
		if n, err := strconv.Atoi(string(r)); err == nil {
			if prevSymbol == "" {
				return "", ErrInvalidString
			}

			if n > 0 {
				_, err := unpacked.WriteString(strings.Repeat(string(prevSymbol), n))
				if err != nil {
					return "", err
				}
			}

			prevSymbol = ""
		} else {
			unpacked.WriteString(prevSymbol)
			prevSymbol = string(r)
		}
	}

	if prevSymbol != "" {
		unpacked.WriteString(prevSymbol)
	}

	return unpacked.String(), nil
}