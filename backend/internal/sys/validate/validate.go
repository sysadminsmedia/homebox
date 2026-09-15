// Package validate provides a wrapper around the go-playground/validator package
package validate

import (
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() { // nolint
	validate = validator.New()

	// Notifier schemes users are allowed to configure. Matching is done on the
	// normalized (lower-cased) scheme rather than a raw prefix: schemes are
	// case-insensitive and shoutrrr lower-cases them before routing, so a raw
	// prefix test disagrees with the service that actually handles the URL.
	allowedNotifierSchemes := map[string]bool{
		"bark":       true,
		"discord":    true,
		"smtp":       true,
		"gotify":     true,
		"googlechat": true,
		"ifttt":      true,
		"join":       true,
		"mattermost": true,
		"matrix":     true,
		"ntfy":       true,
		"opsgenie":   true,
		"pushbullet": true,
		"pushover":   true,
		"rocketchat": true,
		"slack":      true,
		"teams":      true,
		"telegram":   true,
		"zulip":      true,
		"generic":    true,
	}

	err := validate.RegisterValidation("shoutrrr", func(fl validator.FieldLevel) bool {
		str := fl.Field().String()
		if str == "" {
			return false
		}

		scheme, _, ok := splitScheme(str)
		if !ok {
			return false
		}

		// shoutrrr routes on the part before the "+" (generic+http -> generic).
		service, _, _ := strings.Cut(scheme, "+")

		return allowedNotifierSchemes[service]
	})

	if err != nil {
		panic(err)
	}
}

// Check a struct for validation errors and returns any errors the occur. This
// wraps the validate.Struct() function and provides some error wrapping. When
// a validator.ValidationErrors is returned, it is wrapped transformed into a
// FieldErrors array and returned.
func Check(val any) error {
	err := validate.Struct(val)
	if err != nil {
		verrors, ok := err.(validator.ValidationErrors) // nolint - we know it's a validator.ValidationErrors
		if !ok {
			return err
		}

		fields := make(FieldErrors, 0, len(verrors))
		for _, verr := range verrors {
			field := FieldError{
				Field: verr.Field(),
				Error: verr.Error(),
			}

			fields = append(fields, field)
		}
		return fields
	}

	return nil
}
