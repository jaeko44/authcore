package verifier

import (
	"encoding/json"
	"time"

	"authcore.io/authcore/pkg/cryptoutil"
)

const (
	// TOTP represents a TOTP verifier.
	TOTP string = "totp"
)

// TOTPVerifier verifers OTP using TOTP algorithm.
type TOTPVerifier struct {
	MethodName string `json:"method"`
	Secret     string `json:"secret"`
	LastUsed   int64  `json:"last_used,string"`
}

// NewTOTPVerifier returns a new TOTPVerifier instance.
func NewTOTPVerifier(secret string) TOTPVerifier {
	return TOTPVerifier{
		MethodName: TOTP,
		Secret:     secret,
	}
}

// Method returns "totp".
func (v TOTPVerifier) Method() string {
	return TOTP
}

// IsPrimary returns whether this method can be used as the primary authentication.
func (v TOTPVerifier) IsPrimary() bool {
	return false
}

// SkipMFA returns whether this method is sufficient for completing the authentication.
func (v TOTPVerifier) SkipMFA() bool {
	return false
}

// Salt returns nil as TOTP does not require a salt.
func (v TOTPVerifier) Salt() []byte {
	return nil
}

// Request does nothing as TOTP does not a challenge request.
func (v TOTPVerifier) Request(in []byte) (state State, challenge Challenge, err error) {
	return
}

// Verify verifies the incoming TOTP code. Returns true and an updated verifier if the code is
// valid. Otherwise, return false and nil.
func (v TOTPVerifier) Verify(state State, in []byte) (bool, Verifier) {
	answer := string(in)
	a := &cryptoutil.TOTPAuthenticator{
		TotpSecret: v.Secret,
		LastUsedAt: time.Unix(v.LastUsed, 0),
	}
	valid := cryptoutil.ValidateTOTP(answer, a)
	if !valid {
		return false, nil
	}
	v.LastUsed = time.Now().Unix()
	return true, v
}

// TOTPVerifierFromJSON unmarshals a TOTPVerifier from a JSON data.
func TOTPVerifierFromJSON(data []byte) (v Verifier, err error) {
	t := TOTPVerifier{}
	if err = json.Unmarshal(data, &t); err != nil {
		return
	}
	v = t
	return
}
