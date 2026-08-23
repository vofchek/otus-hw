package hw03frequencyanalysis

import (
	"sort"
	"strings"
)

type entry struct {
	key   string
	value uint
}

func Top10(text string) []string {
	partsFrequency := make(map[string]uint)

	for _, part := range strings.Fields(text) {
		_, inMap := partsFrequency[part]
		if !inMap {
			partsFrequency[part] = 0
		}
		partsFrequency[part]++
	}

	if len(partsFrequency) == 0 {
		return make([]string, 0)
	}

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

	result := make([]string, 0, min(10))

	for i := 0; i < 10 && i < len(entries); i++ {
		result = append(result, entries[i].key)
	}

	return result
}
