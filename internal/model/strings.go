package model

import (
	"strconv"
	"strings"
)

func trim(s string) string  { return strings.TrimSpace(s) }
func lower(s string) string { return strings.ToLower(s) }
func itoa(i int) string     { return strconv.Itoa(i) }
func containsFold(s, sub string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(sub))
}
