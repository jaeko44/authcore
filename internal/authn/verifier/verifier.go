package verifier

import (
	"encoding/json"

	"authcore.io/authcore/internal/errors"
)

// Challenge is a challenge for challenge-response verification.
type Challenge []byte

// State represents a states for completing a verification transaction later.
type State []byte

// Verifier defines a factor for verification.
type Verifier interface {
	Method() string
	Salt() []byte
	Request(in []byte) (State, Challenge, error)
	Verify(state State, in []byte) (bool, Verifier)
	IsPrimary() bool
	SkipMFA() bool
}

// Unmarshaller is a function that builds verifier with JSON data.
type Unmarshaller func([]byte) (Verifier, error)

// Factory creates verifier instances.
type Factory struct {
	unmarshallers map[string]Unmarshaller
}

// NewFactory returns a new Factory with defaut verifier set.
func NewFactory() *Factory {
	return &Factory{
		unmarshallers: map[string]Unmarshaller{
			SPAKE2Plus: SPAKE2PlusVerifierFromJSON,
			TOTP:       TOTPVerifierFromJSON,
			BackupCode: BackupCodeVerifierFromJSON,
		},
	}
}

// Register registers a verification method.
func (f *Factory) Register(method string, unmarshaller Unmarshaller) {
	f.unmarshallers[method] = unmarshaller
}

// Unmarshal unmarshals the JSON string to a Verifier
func (f *Factory) Unmarshal(data []byte) (v Verifier, err error) {
	m := make(map[string]string)
	if err = json.Unmarshal(data, &m); err != nil {
		return
	}

	unmarshaller, ok := f.unmarshallers[m["method"]]
	if !ok {
		err = errors.Errorf(errors.ErrorInvalidArgument, "unknonwn password verifier method %v", m["method"])
		return
	}
	return unmarshaller(data)
}
