package dtos

import (
	"fmt"
	"net/mail"
	"regexp"
	"strings"
)

var (
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,32}$`)
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
)

// ValidationErrors collects field-level validation errors.
type ValidationErrors map[string]string

func (ve ValidationErrors) Add(field, message string) {
	ve[field] = message
}

func (ve ValidationErrors) Error() string {
	parts := make([]string, 0, len(ve))
	for field, msg := range ve {
		parts = append(parts, fmt.Sprintf("%s: %s", field, msg))
	}
	return strings.Join(parts, "; ")
}

func (ve ValidationErrors) IsEmpty() bool {
	return len(ve) == 0
}

// AddMap merges another ValidationErrors into this one.
func (ve ValidationErrors) AddMap(other ValidationErrors) {
	for field, msg := range other {
		ve[field] = msg
	}
}

// ValidateEmail checks if the email is valid.
func ValidateEmail(email string) ValidationErrors {
	errs := ValidationErrors{}
	if email == "" {
		errs.Add("email", "is required")
		return errs
	}
	if !emailRegex.MatchString(email) {
		errs.Add("email", "invalid email format")
		return errs
	}
	return errs
}

// ValidatePassword checks password strength.
func ValidatePassword(password string) ValidationErrors {
	errs := ValidationErrors{}
	if password == "" {
		errs.Add("password", "is required")
		return errs
	}
	if len(password) < 8 {
		errs.Add("password", "must be at least 8 characters")
		return errs
	}
	if len(password) > 128 {
		errs.Add("password", "must be at most 128 characters")
		return errs
	}
	return errs
}

// ValidateUsername checks username format.
func ValidateUsername(username string) ValidationErrors {
	errs := ValidationErrors{}
	if username == "" {
		errs.Add("username", "is required")
		return errs
	}
	if !usernameRegex.MatchString(username) {
		errs.Add("username", "must be 3-32 alphanumeric characters or underscores")
		return errs
	}
	return errs
}

// ValidatePositiveInt64 checks that the value is positive.
func ValidatePositiveInt64(value int64, field string) ValidationErrors {
	errs := ValidationErrors{}
	if value <= 0 {
		errs.Add(field, "must be positive")
	}
	return errs
}

// ValidatePositiveInt checks that the value is positive.
func ValidatePositiveInt(value int, field string) ValidationErrors {
	errs := ValidationErrors{}
	if value <= 0 {
		errs.Add(field, "must be positive")
	}
	return errs
}

// ValidatePositiveInt32 checks that the value is positive.
func ValidatePositiveInt32(value int32, field string) ValidationErrors {
	errs := ValidationErrors{}
	if value <= 0 {
		errs.Add(field, "must be positive")
	}
	return errs
}

// ValidateNonEmptyString checks that the string is not empty.
func ValidateNonEmptyString(value, field string) ValidationErrors {
	errs := ValidationErrors{}
	if strings.TrimSpace(value) == "" {
		errs.Add(field, "is required")
	}
	return errs
}

// ValidatePrice checks that the price string is a valid positive number.
func ValidatePrice(price, field string) ValidationErrors {
	errs := ValidationErrors{}
	if strings.TrimSpace(price) == "" {
		errs.Add(field, "is required")
		return errs
	}
	// Basic check: should not start with '-' and should be parseable
	if strings.HasPrefix(price, "-") {
		errs.Add(field, "must be positive")
		return errs
	}
	return errs
}

// ValidateURL checks if the string is a valid URL.
func ValidateURL(url, field string) ValidationErrors {
	errs := ValidationErrors{}
	if url == "" {
		return errs
	}
	_, err := mail.ParseAddress(url)
	if err != nil {
		// Try URL pattern matching
		if matched, _ := regexp.MatchString(`^https?://`, url); matched {
			return errs
		}
		errs.Add(field, "invalid URL format")
	}
	return errs
}
