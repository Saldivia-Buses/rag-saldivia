package intelligence

import "testing"

func TestExtractLifeEvents_DetectsEvents(t *testing.T) {
	messages := []string{
		"Firmé un contrato importante en mayo 2025",
		"Perdí mi trabajo en septiembre 2024",
		"Me mudé a Córdoba en marzo 2023",
		"Hoy hace un lindo día",
	}

	events := ExtractLifeEvents(messages)

	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}

	// Check categories
	categories := map[string]bool{}
	for _, e := range events {
		categories[e.Category] = true
	}
	if !categories["contrato"] {
		t.Error("expected contrato category")
	}
	if !categories["perdida"] {
		t.Error("expected perdida category")
	}
	if !categories["mudanza"] {
		t.Error("expected mudanza category")
	}
}

func TestExtractLifeEvents_ExtractsDates(t *testing.T) {
	messages := []string{
		"Firmé un contrato en mayo 2025",
	}

	events := ExtractLifeEvents(messages)
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	if events[0].Month != 5 {
		t.Errorf("month = %d, want 5", events[0].Month)
	}
	if events[0].Year != 2025 {
		t.Errorf("year = %d, want 2025", events[0].Year)
	}
}

func TestExtractLifeEvents_NoEvents(t *testing.T) {
	messages := []string{
		"Quiero saber sobre mi carrera",
		"Cómo viene el año?",
	}

	events := ExtractLifeEvents(messages)
	if len(events) != 0 {
		t.Errorf("expected 0 events, got %d", len(events))
	}
}

func TestShouldAutoRectify_ThresholdThree(t *testing.T) {
	// Need 3+ events WITH dates
	events := []LifeEvent{
		{Category: "contrato", Month: 5, Year: 2025},
		{Category: "perdida", Month: 9, Year: 2024},
	}
	if ShouldAutoRectify(events) {
		t.Error("should NOT auto-rectify with 2 dated events")
	}

	events = append(events, LifeEvent{Category: "mudanza", Month: 3, Year: 2023})
	if !ShouldAutoRectify(events) {
		t.Error("should auto-rectify with 3 dated events")
	}
}

func TestShouldAutoRectify_IgnoresUndated(t *testing.T) {
	events := []LifeEvent{
		{Category: "contrato", Month: 5, Year: 2025},
		{Category: "perdida", Month: 0, Year: 0}, // no date
		{Category: "mudanza", Month: 3, Year: 2023},
		{Category: "inicio", Month: 0, Year: 0}, // no date
	}
	if ShouldAutoRectify(events) {
		t.Error("should NOT auto-rectify with only 2 dated events")
	}
}

func TestToRectificationEvents_ConvertsCorrectly(t *testing.T) {
	events := []LifeEvent{
		{Description: "Firmé contrato", Category: "contrato", Month: 5, Year: 2025},
		{Description: "Sin fecha", Category: "evento", Month: 0, Year: 0},
		{Description: "Me mudé", Category: "mudanza", Month: 3, Year: 2023},
	}

	rectEvents := ToRectificationEvents(events)
	if len(rectEvents) != 2 {
		t.Fatalf("expected 2 rectification events (skip undated), got %d", len(rectEvents))
	}
	if rectEvents[0].Category != "contrato" {
		t.Errorf("first event category = %q, want contrato", rectEvents[0].Category)
	}
}
