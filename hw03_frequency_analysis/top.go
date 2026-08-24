package hw03frequencyanalysis

import (
	"regexp"
	"sort"
	"strings"
)

var onlyDashRegex = regexp.MustCompile("-{2,}")

type entry struct {
	key   string
	value uint
}

func Top10(text string) []string {
	partsFrequency := make(map[string]uint)

	// считаем повторения слов
	for _, part := range strings.Fields(text) {
		key, isValidKey := tryMakeKey(part)
		if !isValidKey {
			continue
		}

		_, inMap := partsFrequency[key]
		if !inMap {
			partsFrequency[key] = 0
		}
		partsFrequency[key]++
	}

	if len(partsFrequency) == 0 {
		return make([]string, 0)
	}

	// слайс соответсвия строки и кол-ва повторений
	entries := make([]entry, 0, len(partsFrequency))
	for key, value := range partsFrequency {
		entries = append(entries, entry{key, value})
	}

	sort.Slice(entries, func(i int, j int) bool {
		if entries[i].value > entries[j].value {
			return true
		}

		if entries[i].value == entries[j].value {
			return entries[i].key < entries[j].key
		}

		return false
	})

	// результат не больше чем 10 строк
	result := make([]string, 0, min(10))

	for i := 0; i < 10 && i < len(entries); i++ {
		result = append(result, entries[i].key)
	}

	return result
}

// пробуем сделать чистый ключ из представленной строки, путем очистки от спец сиволов
// символы - больше 1 подряд считаем словом
// считаем что нам дали строку без пробела
func tryMakeKey(str string) (string, bool) {
	if onlyDashRegex.MatchString(str) {
		return strings.Trim(str, ",.!?:'\""), true
	}

	cleanKey := strings.Trim(strings.ToLower(str), "-,.!?:'\"")

	if len(cleanKey) == 0 {
		return "", false
	}

	return cleanKey, true
}
