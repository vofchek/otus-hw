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

	// по каждой руне
	for _, r := range packed {
		// пробуем сделать цифру
		n, err := strconv.Atoi(string(r))

		// err пустая, если получили цифру
		// prevSymbol пустой, если на предыдущей итерации получили цифру, тогда рубим выполнение с ошибкой
		if err == nil && prevSymbol == "" {
			return "", ErrInvalidString
		}

		// нашли цифру
		if err == nil {
			// игнорим error WriteString, он всё равно пустой
			unpacked.WriteString(strings.Repeat(prevSymbol, n))
			prevSymbol = ""
			// норм символ, который нужно будет повторять
		} else {
			unpacked.WriteString(prevSymbol)
			prevSymbol = string(r)
		}
	}

	// последний символ не обрабатывается циклом, так что добавляем руками
	if prevSymbol != "" {
		unpacked.WriteString(prevSymbol)
	}

	return unpacked.String(), nil
}
