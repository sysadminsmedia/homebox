package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildFoundContactEmail(t *testing.T) {
	subject, body := buildFoundContactEmail("Camera Bag", "I found this at the park", "finder@example.com")

	assert.Contains(t, subject, "Camera Bag", "subject should contain item name")
	assert.Contains(t, body, "I found this at the park", "body should contain finder message")
	assert.Contains(t, body, "finder@example.com", "body should contain reply address")
}

func TestBuildFoundContactEmail_NoReplyTo(t *testing.T) {
	_, body := buildFoundContactEmail("Camera Bag", "hello", "")
	assert.NotContains(t, body, "Reply to:", "body should omit reply line when empty")
}

func TestBuildFoundContactEmail_EscapesItemName(t *testing.T) {
	_, body := buildFoundContactEmail(`<b>x</b>`, "hello", "")
	assert.NotContains(t, body, "<b>x</b>", "raw item name markup should not appear in body")
	assert.Contains(t, body, "&lt;b&gt;x&lt;/b&gt;", "item name should be HTML-escaped in body")
}

func TestBuildFoundContactEmail_EscapesFinderControlledFields(t *testing.T) {
	_, body := buildFoundContactEmail("x", "<script>alert(1)</script>", "<b>a</b>@x")

	assert.NotContains(t, body, "<script>alert(1)</script>", "raw message markup should not appear in body")
	assert.Contains(t, body, "&lt;script&gt;alert(1)&lt;/script&gt;", "message should be HTML-escaped in body")

	assert.NotContains(t, body, "<b>a</b>@x", "raw reply-to markup should not appear in body")
	assert.Contains(t, body, "&lt;b&gt;a&lt;/b&gt;@x", "reply-to should be HTML-escaped in body")
}
