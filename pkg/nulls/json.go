package nulls

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/pkg/errors"
	// log "github.com/sirupsen/logrus"
)

// JSON adds an implementation for JSON value for MySQL database
// The JSON in the type should be struct type, which the value can be accessed
// directly by the key.
type JSON struct {
	Struct map[string]interface{}
	Valid  bool // Valid is true if JSON is not NULL
}

// Interface implements the nullable interface. It returns nil if
// the byte slice is not valid, otherwise it returns the byte slice value.
func (ns JSON) Interface() interface{} {
	if !ns.Valid {
		return nil
	}
	return ns.Struct
}

// NewJSON returns a new, properly instantiated JSON object.
// Accept struct and JSON format string for the input.
func NewJSON(v interface{}) JSON {
	var b []byte
	var value interface{}
	var m map[string]interface{}
	switch v.(type) {
	case string:
		b = []byte(v.(string))
	case []byte:
		b = v.([]byte)
	case interface{}:
		var err error
		b, err = json.Marshal(v)
		if err != nil {
			return JSON{Struct: nil, Valid: false}
		}
	}
	err := json.Unmarshal(b, &value)
	if err != nil {
		return JSON{Struct: nil, Valid: false}
	}
	if value == nil {
		return JSON{Struct: nil, Valid: false}
	}
	m, ok := value.(map[string]interface{})
	if !ok {
		return JSON{Struct: nil, Valid: false}
	}

	return JSON{Struct: m, Valid: true}
}

// Scan implements the Scanner interface.
func (ns *JSON) Scan(v interface{}) error {
	var value interface{}
	var err error
	switch v := v.(type) {
	case []byte:
		err = json.Unmarshal(v, &value)
		if err != nil {
			ns.Struct = nil
			ns.Valid = false
			return err
		}
		ns.Struct = value.(map[string]interface{})
		ns.Valid = true
		return nil
	case nil:
		ns.Struct = nil
		ns.Valid = false
		return nil
	default:
		ns.Struct = nil
		ns.Valid = false
		return errors.New("undefined type")
	}
}

// Value implements the driver Valuer interface.
func (ns JSON) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	b, err := json.Marshal(ns.Struct)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// String implements the string conversion interface.
func (ns JSON) String() (string, error) {
	if !ns.Valid {
		return "", nil
	}
	b, err := json.Marshal(ns.Struct)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// MarshalJSON implements json.Marshaler.
func (ns JSON) MarshalJSON() ([]byte, error) {
	if !ns.Valid {
		return json.Marshal(nil)
	}
	return json.Marshal(ns.Struct)
}

// UnmarshalJSON will unmarshal a JSON value into
// the propert representation of that value.
func (ns *JSON) UnmarshalJSON(text []byte) error {
	ns.Valid = false
	err := json.Unmarshal(text, &ns.Struct)
	if err != nil {
		return err
	}
	ns.Valid = true
	return nil
}
