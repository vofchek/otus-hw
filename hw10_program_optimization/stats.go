package hw10programoptimization

import (
	"bufio"
	"encoding/json"
	"io"
	"strings"
)

type User struct {
	// ID       int
	// Name     string
	// Username string
	Email string
	// Phone    string
	// Password string
	// Address  string
}

type DomainStat map[string]int

func GetDomainStat(r io.Reader, domain string) (DomainStat, error) {
	result := make(DomainStat)
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)

	line := make([]byte, 0)
	emailDomain := ""
	atIndex := -1
	var user User

	for scanner.Scan() {
		line = scanner.Bytes()

		if err := json.Unmarshal(line, &user); err != nil {
			return nil, err
		}

		atIndex = strings.Index(user.Email, "@")
		if atIndex == -1 {
			continue
		}

		emailDomain = strings.ToLower(user.Email[atIndex+1 : len(user.Email)])
		if !strings.HasSuffix(emailDomain, domain) {
			continue
		}
		result[emailDomain]++
	}

	if scanner.Err() != nil {
		return nil, scanner.Err()
	}

	return result, nil
}
