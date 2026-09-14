package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"slices"
)

// readStrictObject accepts exactly one object containing each named field once.
// Names are case-sensitive; escaped spellings are decoded before duplicate checks.
// Handlers apply the body limit and media type check before calling it, and retain
// their endpoint-specific validation afterward. No caller data enters error text.
func readStrictObject(body io.Reader, destination any, required ...string) error {
	decoder := json.NewDecoder(body)
	opening, err := decoder.Token()
	if err != nil || opening != json.Delim('{') {
		return errors.New("object required")
	}
	fields := make(map[string]json.RawMessage, len(required))
	for decoder.More() {
		key, err := decoder.Token()
		if err != nil {
			return errors.New("invalid field")
		}
		name, ok := key.(string)
		if !ok || !slices.Contains(required, name) {
			return errors.New("unknown field")
		}
		if _, duplicate := fields[name]; duplicate {
			return errors.New("duplicate field")
		}
		var raw json.RawMessage
		if decoder.Decode(&raw) != nil {
			return errors.New("invalid value")
		}
		if string(raw) == "null" {
			return errors.New("null forbidden")
		}
		fields[name] = raw
	}
	closing, err := decoder.Token()
	if err != nil || closing != json.Delim('}') {
		return errors.New("unclosed object")
	}
	if _, err := decoder.Token(); err != io.EOF {
		return errors.New("trailing input")
	}
	if len(fields) != len(required) {
		return errors.New("missing fields")
	}
	// Decode only after the exact field-name/duplicate checks. Decoding straight
	// into a struct would otherwise accept duplicate and case-insensitive names.
	encoded, err := json.Marshal(fields)
	if err != nil {
		return errors.New("invalid object")
	}
	if json.Unmarshal(encoded, destination) != nil {
		return errors.New("invalid field type")
	}
	return nil
}
