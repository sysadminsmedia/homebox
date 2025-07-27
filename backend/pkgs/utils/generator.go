// Package utils
package utils

import (
	"fmt"
	"strings"
	"text/template"
)

// GenerateSubPubConn generates a subscription or publication connection string
func GenerateSubPubConn(pubSubConn string, topic string) (string, error) {
	if strings.Contains(topic, "{{") || strings.Contains(topic, "}}") {
		return "", fmt.Errorf("topic contains template placeholders, which is not allowed")
	}
	builder := &strings.Builder{}
	tmpl, err := template.New("subPubConn").Parse(pubSubConn)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}
	err = tmpl.Execute(builder, map[string]interface{}{
		"Topic": topic,
	})
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}
	return builder.String(), nil
}
