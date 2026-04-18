package spine_test

import (
	"strings"
	"testing"

	"github.com/Camionerou/rag-saldivia/services/app/internal/spine"
)

func TestBuildSubject_SubstitutesPlaceholders(t *testing.T) {
	got, err := spine.BuildSubject("tenant.{slug}.notify.chat.new_message", map[string]string{
		"slug": "saldivia",
	})
	if err != nil {
		t.Fatalf("BuildSubject: %v", err)
	}
	want := "tenant.saldivia.notify.chat.new_message"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestBuildSubject_MultiplePlaceholders(t *testing.T) {
	got, err := spine.BuildSubject("platform.lifecycle.{action}.{tenant_id}", map[string]string{
		"action":    "tenant_created",
		"tenant_id": "saldivia",
	})
	if err != nil {
		t.Fatalf("BuildSubject: %v", err)
	}
	want := "platform.lifecycle.tenant_created.saldivia"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestBuildSubject_RejectsInvalidSlug(t *testing.T) {
	tests := []string{"", "has space", "bad/slash", "with.dot"}
	for _, slug := range tests {
		t.Run(slug, func(t *testing.T) {
			_, err := spine.BuildSubject("tenant.{slug}.notify.x", map[string]string{"slug": slug})
			if err == nil {
				t.Errorf("expected error for slug %q", slug)
			}
		})
	}
}

func TestBuildSubject_UnknownPlaceholder(t *testing.T) {
	_, err := spine.BuildSubject("tenant.{slug}.{unknown}.x", map[string]string{"slug": "saldivia"})
	if err == nil {
		t.Error("expected error for unresolved placeholder")
	}
	if !strings.Contains(err.Error(), "unknown") {
		t.Errorf("error should mention missing key, got %q", err.Error())
	}
}

func TestBuildSubject_RejectsEmptyTemplate(t *testing.T) {
	if _, err := spine.BuildSubject("", map[string]string{}); err == nil {
		t.Error("expected error for empty template")
	}
}

func TestBuildSubject_AcceptsDotsInTemplate(t *testing.T) {
	got, err := spine.BuildSubject("tenant.{slug}.notify.chat.new_message", map[string]string{
		"slug": "x",
	})
	if err != nil {
		t.Fatalf("BuildSubject: %v", err)
	}
	if got != "tenant.x.notify.chat.new_message" {
		t.Errorf("unexpected result: %q", got)
	}
}

func TestValidateSubject_AcceptsCanonical(t *testing.T) {
	valid := []string{
		"tenant.saldivia.notify.chat.new_message",
		"platform.lifecycle.tenant_created",
		"dlq.NOTIFICATIONS.tenant.saldivia.notify.chat.new_message",
	}
	for _, s := range valid {
		t.Run(s, func(t *testing.T) {
			if err := spine.ValidateSubject(s); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateSubject_RejectsMalformed(t *testing.T) {
	invalid := []string{
		"",
		".leading.dot",
		"trailing.dot.",
		"double..dot",
		"has space.in.segment",
		"has/slash.in.segment",
	}
	for _, s := range invalid {
		t.Run(s, func(t *testing.T) {
			if err := spine.ValidateSubject(s); err == nil {
				t.Errorf("expected error for %q", s)
			}
		})
	}
}
