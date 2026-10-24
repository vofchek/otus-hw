package hw10programoptimization

import (
	"bufio"
	"encoding/json"
	"fmt"
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
	u, err := getUsers(r)
	if err != nil {
		return nil, fmt.Errorf("get users error: %w", err)
	}
	return countDomains(u, domain)
}

type users []User

func getUsers(r io.Reader) (result users, err error) {
	result = make(users, 0, 100_000)

	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)

	line := make([]byte, 0)
	for scanner.Scan() {
		line = scanner.Bytes()

		var user User
		if err = json.Unmarshal(line, &user); err != nil {
			return
		}
		result = append(result, user)
	}

	err = scanner.Err()

	return
}

func countDomains(u users, domain string) (DomainStat, error) {
	result := make(DomainStat)
	emailDomain := ""
	atIndex := -1

	for _, user := range u {
		atIndex = strings.Index(user.Email, "@")
		if atIndex == -1 {
			continue
		}

		emailDomain = strings.ToLower(user.Email[atIndex+1 : len(user.Email)])
		if !strings.Contains(emailDomain, domain) {
			continue
		}
		result[emailDomain]++
	}
	return result, nil
}
