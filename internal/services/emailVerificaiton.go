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

func HandleEmailError(err error) (int, string) {
    switch err {
    case ErrInvalidFormat:
        return 400, "Invalid email format"
    case ErrInvalidDomain:
        return 400, "Email domain does not exist or cannot receive mail"
    default:
        return 500, "Verification service error"
    }
}