// Package validate is a tiny, explicit input validator for handlers/forms: you
// accumulate per-field error messages as you check, with NO reflection and NO
// struct tags — nothing hidden, you read the checks. Each domain composes only
// the rules it needs. It works on strings (form values); parse numbers/dates in
// the handler and validate the result.
//
// It's a leaf utility (like respond or password), not a Module: handlers import
// and call it; it registers nothing on the App.
package validate

import (
	"net/mail"
	"strings"
	"unicode/utf8"
)

// Errors maps a field name to its error message. Empty means valid. Its
// underlying type is map[string]string, so it drops straight into a template's
// page data (e.g. {{.Errors.email}}).
type Errors map[string]string

// New returns an empty error set ready to accumulate.
func New() Errors { return Errors{} }

// OK reports whether every field passed.
func (e Errors) OK() bool { return len(e) == 0 }

// Add records msg for field, keeping the FIRST message set for that field, so
// the earliest (most important) rule wins and later checks don't overwrite it.
func (e Errors) Add(field, msg string) {
	if _, seen := e[field]; !seen {
		e[field] = msg
	}
}

// Required fails if value is empty after trimming whitespace.
func (e Errors) Required(field, value, msg string) {
	if strings.TrimSpace(value) == "" {
		e.Add(field, msg)
	}
}

// MinLen fails if the trimmed value has fewer than n characters (runes, so
// accents and emoji count as one).
func (e Errors) MinLen(field, value string, n int, msg string) {
	if utf8.RuneCountInString(strings.TrimSpace(value)) < n {
		e.Add(field, msg)
	}
}

// MaxLen fails if value has more than n characters (runes).
func (e Errors) MaxLen(field, value string, n int, msg string) {
	if utf8.RuneCountInString(value) > n {
		e.Add(field, msg)
	}
}

// Email fails if value is not a parseable email address (net/mail, stdlib).
func (e Errors) Email(field, value, msg string) {
	if _, err := mail.ParseAddress(value); err != nil {
		e.Add(field, msg)
	}
}
