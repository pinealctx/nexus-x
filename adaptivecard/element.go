package adaptivecard

import "encoding/json"

// Element is the interface implemented by all card body elements.
type Element interface {
	elementType() string
}

// ElementBase holds common fields shared by all elements (version 1.0–1.5).
// Embed this in concrete element types.
type ElementBase struct {
	Height    BlockElementHeight `json:"height,omitempty"`
	Separator bool               `json:"separator,omitempty"`
	Spacing   Spacing            `json:"spacing,omitempty"`
	ID        string             `json:"id,omitempty"`
	IsVisible *bool              `json:"isVisible,omitempty"`
	Requires  map[string]string  `json:"requires,omitempty"`
}

// elementJSON is a helper for marshaling elements with a "type" discriminator.
// Concrete types call json.Marshal on an alias to avoid recursion.

// unmarshalElement decodes a JSON object into the correct Element type
// by inspecting the "type" field.
func unmarshalElement(raw json.RawMessage) (Element, error) {
	var probe struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(raw, &probe); err != nil {
		return nil, err
	}
	var el Element
	switch probe.Type {
	case "TextBlock":
		el = &TextBlock{}
	case "Image":
		el = &Image{}
	case "RichTextBlock":
		el = &RichTextBlock{}
	case "Media":
		el = &Media{}
	case "Container":
		el = &Container{}
	case "ColumnSet":
		el = &ColumnSet{}
	case "FactSet":
		el = &FactSet{}
	case "ImageSet":
		el = &ImageSet{}
	case "ActionSet":
		el = &ActionSet{}
	case "Table":
		el = &Table{}
	case "Input.Text":
		el = &InputText{}
	case "Input.Number":
		el = &InputNumber{}
	case "Input.Date":
		el = &InputDate{}
	case "Input.Time":
		el = &InputTime{}
	case "Input.Toggle":
		el = &InputToggle{}
	case "Input.ChoiceSet":
		el = &InputChoiceSet{}
	default:
		// Unknown element: preserve as raw JSON.
		return &RawElement{RawJSON: raw}, nil
	}
	if err := json.Unmarshal(raw, el); err != nil {
		return nil, err
	}
	return el, nil
}

// unmarshalSlice is a generic helper that decodes a JSON array of
// polymorphic items using the supplied per-item decoder.
func unmarshalSlice[T any](raw json.RawMessage, decode func(json.RawMessage) (T, error)) ([]T, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	var arr []json.RawMessage
	if err := json.Unmarshal(raw, &arr); err != nil {
		return nil, err
	}
	out := make([]T, 0, len(arr))
	for _, r := range arr {
		v, err := decode(r)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, nil
}

// unmarshalElements decodes a JSON array of elements.
func unmarshalElements(raw json.RawMessage) ([]Element, error) {
	return unmarshalSlice(raw, unmarshalElement)
}

// RawElement preserves an unknown element type as raw JSON.
type RawElement struct {
	RawJSON json.RawMessage
}

func (r *RawElement) elementType() string { return "unknown" }

func (r *RawElement) MarshalJSON() ([]byte, error) {
	return r.RawJSON, nil
}

// InputBase holds common fields shared by all input elements (version 1.3+).
type InputBase struct {
	ElementBase
	ErrorMessage string `json:"errorMessage,omitempty"`
	IsRequired   *bool  `json:"isRequired,omitempty"`
	Label        string `json:"label,omitempty"`
}
