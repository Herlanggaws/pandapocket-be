package database

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// JSONRaw stores JSON as TEXT and scans both string and []byte driver values.
// Needed because Go 1.25 aliases json.RawMessage to jsontext.Value, which cannot
// Scan PostgreSQL TEXT returned as string.
type JSONRaw []byte

func (j *JSONRaw) Scan(value interface{}) error {
	if value == nil {
		*j = JSONRaw("{}")
		return nil
	}
	switch v := value.(type) {
	case string:
		*j = JSONRaw(v)
	case []byte:
		*j = append(JSONRaw(nil), v...)
	default:
		return fmt.Errorf("JSONRaw: unsupported Scan type %T", value)
	}
	return nil
}

func (j JSONRaw) Value() (driver.Value, error) {
	if len(j) == 0 {
		return "{}", nil
	}
	return string(j), nil
}

func (j JSONRaw) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("{}"), nil
	}
	return json.RawMessage(j).MarshalJSON()
}

func (j *JSONRaw) UnmarshalJSON(data []byte) error {
	if j == nil {
		return fmt.Errorf("JSONRaw: UnmarshalJSON on nil pointer")
	}
	*j = append((*j)[0:0], data...)
	return nil
}
