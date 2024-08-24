package utils

import "go/format"

func FormatCode(rawCode string) (string, error) {
	formated, formatErr := format.Source([]byte(rawCode))
	if formatErr != nil {
		return "", formatErr
	}
	return string(formated), nil
}
