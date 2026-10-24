package contact

import (
	"context"
	"testing"
)

func TestParseEmailContent_StripsBOM(t *testing.T) {
	content := string(rune(0xFEFF)) + "email,attributes\na@example.com,\"{\"\"name\"\":\"\"Nguyễn\"\"}\"\n"
	contacts := parseEmailContent(context.Background(), content, ImportTypeFile)
	if len(contacts) != 1 {
		t.Fatalf("expected 1 contact, got %d", len(contacts))
	}
	if contacts[0].Email != "a@example.com" || contacts[0].Attribs["name"] != "Nguyễn" {
		t.Errorf("unexpected contact %+v", contacts[0])
	}
}
