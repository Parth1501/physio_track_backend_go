package whatsapp

import (
	"errors"
	"strings"
)

var (
	ErrNoPhone      = errors.New("patient has no phone number")
	ErrInvalidPhone = errors.New("phone number is not a valid Indian mobile number")
)

// NormalizeIndian converts a stored phone number into WhatsApp's international format (91XXXXXXXXXX).
// Accepts spaces, dashes, brackets, a +91/91 prefix or a leading trunk 0.
func NormalizeIndian(raw string) (string, error) {
	var b strings.Builder
	for _, r := range raw {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	digits := b.String()
	if digits == "" {
		return "", ErrNoPhone
	}
	switch {
	case len(digits) == 12 && strings.HasPrefix(digits, "91"):
		digits = digits[2:]
	case len(digits) == 11 && strings.HasPrefix(digits, "0"):
		digits = digits[1:]
	}
	if len(digits) != 10 || digits[0] < '6' {
		return "", ErrInvalidPhone
	}
	return "91" + digits, nil
}
