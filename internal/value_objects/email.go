package value_objects

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// Email represents an email address value object
type Email struct {
	value string
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// NewEmail creates a new Email value object
func NewEmail(value string) (*Email, error) {
	value = strings.TrimSpace(strings.ToLower(value))

	if value == "" {
		return nil, errors.New("email cannot be empty")
	}

	if !emailRegex.MatchString(value) {
		return nil, fmt.Errorf("invalid email format: %s", value)
	}

	return &Email{value: value}, nil
}

// String returns the string representation of the email
func (e Email) String() string {
	return e.value
}

// Value returns the email value
func (e Email) Value() string {
	return e.value
}

// Equals checks if two emails are equal
func (e Email) Equals(other interface{}) bool {
	otherEmail, ok := other.(*Email)
	if !ok {
		return false
	}
	return e.value == otherEmail.value
}

// Domain returns the domain part of the email
func (e Email) Domain() string {
	parts := strings.Split(e.value, "@")
	if len(parts) == 2 {
		return parts[1]
	}
	return ""
}

// Username returns the username part of the email
func (e Email) Username() string {
	parts := strings.Split(e.value, "@")
	if len(parts) == 2 {
		return parts[0]
	}
	return ""
}

// MarshalJSON implements json.Marshaler
func (e Email) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, e.value)), nil
}

// UnmarshalJSON implements json.Unmarshaler
func (e *Email) UnmarshalJSON(data []byte) error {
	value := strings.Trim(string(data), `"`)
	email, err := NewEmail(value)
	if err != nil {
		return err
	}
	*e = *email
	return nil
}