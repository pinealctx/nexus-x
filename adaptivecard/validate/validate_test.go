package validate_test

import (
	"testing"

	ac "github.com/pinealctx/nexus-x/adaptivecard"
	"github.com/pinealctx/nexus-x/adaptivecard/validate"
)

func TestSchemaValidCard(t *testing.T) {
	card := ac.NewCard().AddBody(ac.NewTextBlock("Hello"))
	j, err := card.JSON()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := validate.ValidateJSON(j); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestSchemaValidCardWithActions(t *testing.T) {
	card := ac.NewCard().
		AddBody(ac.NewTextBlock("Title")).
		AddAction(ac.NewActionSubmit("OK", nil))
	j, err := card.JSON()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := validate.ValidateJSON(j); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestSchemaInvalidJSON(t *testing.T) {
	err := validate.ValidateJSON(`{not valid json}`)
	if !validate.IsValidationError(err, "invalid_json") {
		t.Errorf("expected invalid_json, got %v", err)
	}
}

func TestSchemaInvalidType(t *testing.T) {
	err := validate.ValidateJSON(`{"type":"SomethingElse","version":"1.5","body":[{"type":"TextBlock","text":"hi"}]}`)
	if !validate.IsValidationError(err, "schema_violation") {
		t.Errorf("expected schema_violation, got %v", err)
	}
}

func TestSchemaUnknownElement(t *testing.T) {
	jsonStr := `{
		"type": "AdaptiveCard",
		"version": "1.5",
		"body": [{"type": "CustomWidget", "text": "hi"}]
	}`
	err := validate.ValidateJSON(jsonStr)
	if !validate.IsValidationError(err, "schema_violation") {
		t.Errorf("expected schema_violation, got %v", err)
	}
}

func TestSchemaUnknownAction(t *testing.T) {
	jsonStr := `{
		"type": "AdaptiveCard",
		"version": "1.5",
		"body": [{"type": "TextBlock", "text": "hi"}],
		"actions": [{"type": "Action.Custom", "title": "x"}]
	}`
	err := validate.ValidateJSON(jsonStr)
	if !validate.IsValidationError(err, "schema_violation") {
		t.Errorf("expected schema_violation, got %v", err)
	}
}

func TestEmptyCard(t *testing.T) {
	err := validate.ValidateJSON(`{"type":"AdaptiveCard","version":"1.5","body":[]}`)
	if !validate.IsValidationError(err, "empty_card") {
		t.Errorf("expected empty_card, got %v", err)
	}
}

func TestEmptyCardNoFields(t *testing.T) {
	// Schema requires "type" and "version", but no body/actions.
	// This should fail with empty_card since body/actions are absent.
	err := validate.ValidateJSON(`{"type":"AdaptiveCard","version":"1.5"}`)
	if !validate.IsValidationError(err, "empty_card") {
		t.Errorf("expected empty_card, got %v", err)
	}
}

func TestActionsOnlyCard(t *testing.T) {
	card := ac.NewCard().AddAction(ac.NewActionSubmit("OK", nil))
	j, err := card.JSON()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := validate.ValidateJSON(j); err != nil {
		t.Fatalf("actions-only card should be valid, got %v", err)
	}
}

func TestValidateCardTyped(t *testing.T) {
	card := ac.NewCard().AddBody(ac.NewTextBlock("Hello"))
	if err := validate.ValidateCard(card); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestValidateCardEmpty(t *testing.T) {
	card := ac.NewCard()
	err := validate.ValidateCard(card)
	if err == nil {
		t.Fatal("expected error for empty card")
	}
	if !validate.IsValidationError(err, "empty_card") {
		t.Errorf("expected empty_card, got %v", err)
	}
}

func TestValidateCardNil(t *testing.T) {
	if err := validate.ValidateCard(nil); err == nil {
		t.Fatal("expected error for nil card")
	}
}

func TestComplexNestedCard(t *testing.T) {
	card := ac.NewCard().
		AddBody(ac.NewContainer(
			ac.NewColumnSet(
				ac.NewColumn(
					ac.NewTextBlock("Left"),
				).SetWidth("auto"),
				ac.NewColumn(
					ac.NewImage("https://example.com/img.png"),
				).SetWidth("stretch"),
			),
		)).
		AddBody(ac.NewTable().
			AddColumn(1).
			AddColumn(2).
			AddTextRow("A", "B")).
		AddBody(ac.NewFactSet().
			AddFact("Key", "Val")).
		AddBody(ac.NewActionSet().
			AddAction(ac.NewActionSubmit("Submit", nil))).
		AddAction(ac.NewActionOpenURL("Visit", "https://example.com"))

	if err := validate.ValidateCard(card); err != nil {
		t.Fatalf("complex nested card should be valid, got %v", err)
	}
}
