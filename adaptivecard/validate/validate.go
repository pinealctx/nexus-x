// Package validate provides Adaptive Card JSON validation using the official
// Adaptive Card JSON Schema and Nexus-specific business rules.
//
// Two validation layers:
//   1. JSON Schema validation (structural correctness against official schema)
//   2. Nexus business rules (e.g., empty card rejection)
package validate

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"

	ac "github.com/pinealctx/nexus-x/adaptivecard"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

//go:embed adaptive-card.json
var schemaBytes []byte

// compiledSchema is the compiled Adaptive Card JSON Schema.
// Compiled once at init; thread-safe for concurrent Validate calls.
var compiledSchema = mustCompileSchema()

func mustCompileSchema() *jsonschema.Schema {
	sch, err := jsonschema.CompileString("adaptive-card.json", string(schemaBytes))
	if err != nil {
		panic(fmt.Sprintf("adaptivecard/validate: compile schema: %v", err))
	}
	return sch
}

// --- Error types ---

// ValidationError describes a card validation failure.
type ValidationError struct {
	Code    string // "invalid_json" | "schema_violation" | "empty_card"
	Message string
	Path    string // JSON Pointer (non-empty for schema_violation)
}

func (e *ValidationError) Error() string { return "adaptivecard/validate: " + e.Message }

// Predefined sentinel errors.
var (
	ErrInvalidJSON = &ValidationError{
		Code:    "invalid_json",
		Message: "invalid JSON syntax",
	}
	ErrEmptyCard = &ValidationError{
		Code:    "empty_card",
		Message: "card must have at least one body element or action",
	}
)

// IsValidationError checks if err is a *ValidationError with the given code.
func IsValidationError(err error, code string) bool {
	var ve *ValidationError
	return errors.As(err, &ve) && ve.Code == code
}

// --- Public API ---

// ValidateJSON validates a raw JSON string as an Adaptive Card.
// It runs JSON Schema validation followed by Nexus business rules.
func ValidateJSON(jsonStr string) error {
	// Layer 1: JSON syntax check.
	var v any
	if err := json.Unmarshal([]byte(jsonStr), &v); err != nil {
		return ErrInvalidJSON
	}

	// Layer 2: JSON Schema validation against the official Adaptive Card schema.
	if err := compiledSchema.Validate(v); err != nil {
		return schemaError(err)
	}

	// Layer 3: Nexus business rules.
	var card ac.Card
	if err := json.Unmarshal([]byte(jsonStr), &card); err != nil {
		return ErrInvalidJSON
	}
	return validateBusiness(&card)
}

// ValidateCard validates a typed *adaptivecard.Card.
// It marshals the card to JSON and delegates to ValidateJSON.
func ValidateCard(card *ac.Card) error {
	if card == nil {
		return ErrInvalidJSON
	}
	data, err := json.Marshal(card)
	if err != nil {
		return ErrInvalidJSON
	}
	return ValidateJSON(string(data))
}

// --- Internal helpers ---

// schemaError converts a jsonschema.ValidationError into a ValidationError.
// It extracts the leaf error for the most specific message and path.
func schemaError(err error) *ValidationError {
	jsErr, ok := err.(*jsonschema.ValidationError)
	if !ok {
		return &ValidationError{
			Code:    "schema_violation",
			Message: err.Error(),
		}
	}

	// Walk to the leaf error for the most specific information.
	leaf := jsErr
	for len(leaf.Causes) > 0 {
		leaf = leaf.Causes[0]
	}

	msg := leaf.Message
	if msg == "" {
		msg = jsErr.Message
	}

	return &ValidationError{
		Code:    "schema_violation",
		Message: msg,
		Path:    leaf.InstanceLocation,
	}
}

// validateBusiness checks Nexus-specific business rules.
func validateBusiness(card *ac.Card) error {
	if len(card.Body) == 0 && len(card.Actions) == 0 {
		return ErrEmptyCard
	}
	return nil
}
