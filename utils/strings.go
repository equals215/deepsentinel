package utils

import (
	"strings"
	"unicode"
)

// CleanString removes all non-graphic characters from a string
func CleanString(dirty string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsGraphic(r) {
			return r
		}
		return -1
	}, dirty)
}
