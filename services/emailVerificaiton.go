package services

import (
    "errors"
    "net"
    "regexp"
    "strings"
)

var (
	ErrInvalidFormat = errors.New("invalid email format")
	ErrInvalidDomain = errors.New("email domain does not exist")
)

var emailRegex = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)

func ValidateEmail(email string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	if !emailRegex.MatchString(email) {
		return ErrInvalidFormat
	}
	at := strings.LastIndex(email, "@")
	domain := email[at+1:]
	mx, err := net.LookupMX(domain)
	if err != nil || len(mx) == 0 {
		return ErrInvalidDomain
	}

	return nil
}