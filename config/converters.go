package config

import (
	"regexp"
	"strings"
)

var regex = regexp.MustCompile("([A-Z][a-z]+)([A-Z]+[a-z]*)")

func CamelToSnake(field string) string {
	result := regex.ReplaceAllString(field, "${1}_${2}")
	return strings.ToLower(result)
}
