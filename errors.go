package validator

import (
	"fmt"
	"strings"
)

type Errors map[string][]string

func (e Errors) Add(field, message string) {
	e[field] = append(e[field], message)
}
func (e Errors) Get(field string) []string {
	if messages, ok := e[field]; ok {
		return messages
	}
	return []string{}
}
func (e Errors) First(field string) string {
	if messages := e.Get(field); len(messages) > 0 {
		return messages[0]
	}
	return ""
}
func (e Errors) Has(field string) bool {
	_, ok := e[field]
	return ok
}
func (e Errors) Error() string {
	var builder strings.Builder
	builder.WriteString("validation failed with the following errors:\n")
	for field, messages := range e {
		builder.WriteString(fmt.Sprintf("  - field '%s': %s\n", field, strings.Join(messages, ", ")))
	}
	return builder.String()
}
