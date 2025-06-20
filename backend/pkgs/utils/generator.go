// Package utils
package utils

import (
	"strings"
	"text/template"
)

// GenerateSubPubConn generates a subscription or publication connection string
func GenerateSubPubConn(pubSubConn string, topic string) string {
	builder := &strings.Builder{}
	tmpl, err := template.New("subPubConn").Parse(pubSubConn)
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(builder, map[string]string{
		"Topic": topic,
	})
	return builder.String()
}
