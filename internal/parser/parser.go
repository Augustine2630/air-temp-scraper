package parser

import (
	"encoding/json"
	"fmt"
	"io"
)

// SensorReading represents a single temperature sensor entry from the JSON source.
type SensorReading struct {
	Name     string
	Key      string
	Type     string
	Value    string
	Quantity *float64
	Unit     string
}

// rawEntry mirrors the JSON structure for a single sensor object.
type rawEntry struct {
	Key      string   `json:"key"`
	Type     string   `json:"type"`
	Value    string   `json:"value"`
	Quantity *float64 `json:"quantity"`
	Unit     string   `json:"unit"`
}

// Parse streams JSON from r and calls fn for each sensor reading.
// It avoids loading the full document into memory.
func Parse(r io.Reader, fn func(SensorReading)) error {
	dec := json.NewDecoder(r)

	// Expect opening '{'
	tok, err := dec.Token()
	if err != nil {
		return fmt.Errorf("read opening token: %w", err)
	}
	if delim, ok := tok.(json.Delim); !ok || delim != '{' {
		return fmt.Errorf("expected '{', got %v", tok)
	}

	var entry rawEntry

	for dec.More() {
		// Read sensor name key
		nameTok, err := dec.Token()
		if err != nil {
			return fmt.Errorf("read sensor name: %w", err)
		}
		name, ok := nameTok.(string)
		if !ok {
			return fmt.Errorf("expected sensor name string, got %T", nameTok)
		}

		// Reset entry to avoid stale data from previous iteration
		entry = rawEntry{}
		if err := dec.Decode(&entry); err != nil {
			return fmt.Errorf("decode entry %q: %w", name, err)
		}

		fn(SensorReading{
			Name:     name,
			Key:      entry.Key,
			Type:     entry.Type,
			Value:    entry.Value,
			Quantity: entry.Quantity,
			Unit:     entry.Unit,
		})
	}

	// Consume closing '}'
	if _, err := dec.Token(); err != nil {
		return fmt.Errorf("read closing token: %w", err)
	}

	return nil
}
