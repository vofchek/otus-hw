package hw10programoptimization

import (
	"bufio"
	"encoding/json"
	"io"
	"strings"
)

type User struct {
	Email string
}

type DomainStat map[string]int

func GetDomainStat(r io.Reader, domain string) (DomainStat, error) {
	result := make(DomainStat)
	scanner := bufio.NewScanner(r)
	suffix := "." + domain

	for scanner.Scan() {
		line := scanner.Bytes()

		var user User
		if err := json.Unmarshal(line, &user); err != nil {
			return nil, err
		}

		atIndex := strings.Index(user.Email, "@")
		if atIndex == -1 {
			continue
		}

		emailDomain := strings.ToLower(user.Email[atIndex+1:])
		if strings.HasSuffix(emailDomain, suffix) {
			result[emailDomain]++
		}
	}

	if scanner.Err() != nil {
		return nil, scanner.Err()
	}

	return result, nil
}
