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
	unpackedPart := ""
	escaping := false

	// по каждой руне
	for _, r := range packed {
		if !escaping && r == '\\' {
			escaping = true
			unpackedPart = prevSymbol
			prevSymbol = ""
			unpacked.WriteString(unpackedPart)
			continue
		}

		// пробуем сделать цифру
		number, notNumber := strconv.Atoi(string(r))

		if escaping && notNumber != nil && r != '\\' {
			return "", ErrInvalidString
		}

		// err пустая, если получили цифру
		// prevSymbol пустой, если на предыдущей итерации получили цифру, тогда рубим выполнение с ошибкой
		if !escaping && notNumber == nil && prevSymbol == "" {
			return "", ErrInvalidString
		}

		// нашли цифру
		if !escaping && notNumber == nil {
			// игнорим error WriteString, он всё равно пустой
			unpackedPart = strings.Repeat(prevSymbol, number)
			prevSymbol = ""
			// норм символ, который нужно будет повторять
		} else {
			unpackedPart = prevSymbol
			prevSymbol = string(r)
		}

		unpacked.WriteString(unpackedPart)
		escaping = false
	}

	// последний символ не обрабатывается циклом, так что добавляем руками
	if prevSymbol != "" {
		unpacked.WriteString(prevSymbol)
	}

	return unpacked.String(), nil
}
