package client

import "encoding/json"

// marshalCard converts a card object to JSON bytes.
func marshalCard(card any) ([]byte, error) {
	switch v := card.(type) {
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	case json.RawMessage:
		return v, nil
	default:
		return json.Marshal(card)
	}
}
