package batch_mail

import (
	"context"
	"testing"

	"billionmail-core/internal/model/entity"
)

func TestProcessMailContentAndSubject_SubjectOverride(t *testing.T) {
	tpl := &entity.ApiTemplates{Unsubscribe: 0}

	_, subject := processMailContentAndSubject(context.Background(), "<p>hi</p>", "Template subject", tpl, entity.Contact{}, ApiMailLog{Subject: "  Order #42 shipped  "})
	if subject != "Order #42 shipped" {
		t.Errorf("override not applied, got %q", subject)
	}

	_, subject = processMailContentAndSubject(context.Background(), "<p>hi</p>", "Template subject", tpl, entity.Contact{}, ApiMailLog{})
	if subject != "Template subject" {
		t.Errorf("template subject should be kept, got %q", subject)
	}
}
