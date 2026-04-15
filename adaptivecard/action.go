package adaptivecard

import "encoding/json"

// Action is the interface implemented by all card actions.
type Action interface {
	actionType() string
}

// ActionBase holds common fields shared by all actions (version 1.0–1.5).
type ActionBase struct {
	Title     string            `json:"title,omitempty"`
	IconURL   string            `json:"iconUrl,omitempty"`
	ID        string            `json:"id,omitempty"`
	Style     ActionStyle       `json:"style,omitempty"`
	Tooltip   string            `json:"tooltip,omitempty"`
	IsEnabled *bool             `json:"isEnabled,omitempty"`
	Mode      ActionMode        `json:"mode,omitempty"`
	Requires  map[string]string `json:"requires,omitempty"`
}

// unmarshalAction decodes a JSON object into the correct Action type.
func unmarshalAction(raw json.RawMessage) (Action, error) {
	var probe struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(raw, &probe); err != nil {
		return nil, err
	}
	var act Action
	switch probe.Type {
	case "Action.OpenUrl":
		act = &ActionOpenURL{}
	case "Action.Submit":
		act = &ActionSubmit{}
	case "Action.ShowCard":
		act = &ActionShowCard{}
	case "Action.ToggleVisibility":
		act = &ActionToggleVisibility{}
	case "Action.Execute":
		act = &ActionExecute{}
	case "Action.OpenMiniApp":
		act = &ActionOpenMiniApp{}
	default:
		return &RawAction{RawJSON: raw}, nil
	}
	if err := json.Unmarshal(raw, act); err != nil {
		return nil, err
	}
	return act, nil
}

// unmarshalActions decodes a JSON array of actions.
func unmarshalActions(raw json.RawMessage) ([]Action, error) {
	return unmarshalSlice(raw, unmarshalAction)
}

// unmarshalOptionalAction decodes a single optional Action from a raw
// JSON field extracted from a map. Returns nil if raw is empty/null.
func unmarshalOptionalAction(raw json.RawMessage) (Action, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	return unmarshalAction(raw)
}

// interfaceFields extracts named keys from a JSON object, deletes them,
// and returns the remaining bytes for standard struct decoding.
func extractFields(data []byte, keys ...string) (map[string]json.RawMessage, []byte, error) {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(data, &obj); err != nil {
		return nil, nil, err
	}
	extracted := make(map[string]json.RawMessage, len(keys))
	for _, k := range keys {
		if v, ok := obj[k]; ok {
			extracted[k] = v
			delete(obj, k)
		}
	}
	rest, err := json.Marshal(obj)
	if err != nil {
		return nil, nil, err
	}
	return extracted, rest, nil
}

// RawAction preserves an unknown action type as raw JSON.
type RawAction struct {
	RawJSON json.RawMessage
}

func (r *RawAction) actionType() string { return "unknown" }

func (r *RawAction) MarshalJSON() ([]byte, error) {
	return r.RawJSON, nil
}
