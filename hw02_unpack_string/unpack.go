package hw02unpackstring

import (
	"errors"
	"strconv"
	"strings"
)

var ErrInvalidString = errors.New("invalid string")

func Unpack(packed string) (string, error) {
	// проверка на пустую строку, чтобы сразу выйти и ничего не обрабатывать
	if packed == "" {
		return "", nil
	}

	unpacked := strings.Builder{}
	// предудущий символ
	prevSymbol := ""
	// часть строки, которую будем добавлять в unpacked
	unpackedPart := ""
	// режим экранирования
	escaping := false

	// по каждой руне
	for _, r := range packed {
		// режим экранирования
		if escaping {
			// только цифры и \
			if (r < '0' || r > '9') && r != '\\' {
				return "", ErrInvalidString
			}

			unpackedPart = prevSymbol
			prevSymbol = string(r)
			escaping = false
		} else {
			switch {
			// включение экранирования
			case r == '\\':
				escaping = true
				unpackedPart = prevSymbol
				prevSymbol = ""
			// цифры
			case r >= '0' && r <= '9':
				if prevSymbol == "" {
					return "", ErrInvalidString
				}

				// пробуем сделать цифру, игнори ошибку из-за проверки в case
				number, _ := strconv.Atoi(string(r))
				unpackedPart = strings.Repeat(prevSymbol, number)
				prevSymbol = ""
			default:
				unpackedPart = prevSymbol
				prevSymbol = string(r)
			}
		}

		unpacked.WriteString(unpackedPart)
	}

	// последний символ не обрабатывается циклом, так что добавляем руками
	if prevSymbol != "" {
		unpacked.WriteString(prevSymbol)
	}

	return unpacked.String(), nil
}
