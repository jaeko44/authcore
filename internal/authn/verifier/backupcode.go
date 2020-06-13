package verifier

import (
	"encoding/json"

	"authcore.io/authcore/pkg/cryptoutil"
)

const (
	// BackupCode represents a backup code verifier.
	BackupCode string = "backup_code"
)

// BackupCodeVerifier verifers backup code.
type BackupCodeVerifier struct {
	MethodName   string `json:"method"`
	Secret       string `json:"secret"`
	UsedCodeMask int64  `json:"used_code_mask,string"`
}

// NewBackupCodeVerifier returns a new BackupCodeVerifier instance.
func NewBackupCodeVerifier(secret string) BackupCodeVerifier {
	return BackupCodeVerifier{
		MethodName: BackupCode,
		Secret:     secret,
	}
}

// Method returns "backup_code".
func (v BackupCodeVerifier) Method() string {
	return BackupCode
}

// IsPrimary returns whether this method can be used as the primary authentication.
func (v BackupCodeVerifier) IsPrimary() bool {
	return false
}

// SkipMFA returns whether this method is sufficient for completing the authentication.
func (v BackupCodeVerifier) SkipMFA() bool {
	return false
}

// Salt returns nil as backup code does not require a salt.
func (v BackupCodeVerifier) Salt() []byte {
	return nil
}

// Request does nothing as backup code does not a challenge request.
func (v BackupCodeVerifier) Request(in []byte) (state State, challenge Challenge, err error) {
	return
}

// Verify verifies the incoming BackupCode code. Returns true and an updated verifier if the code is
// valid. Otherwise, return false and nil.
func (v BackupCodeVerifier) Verify(state State, in []byte) (bool, Verifier) {
	answer := string(in)
	usedCodeMask, err := cryptoutil.ValidateBackupCodes(answer, v.Secret, v.UsedCodeMask, uint64(10))
	if err != nil {
		return false, nil
	}
	v.UsedCodeMask = usedCodeMask
	return true, v
}

// BackupCodeVerifierFromJSON unmarshals a BackupCodeVerifier from a JSON data.
func BackupCodeVerifierFromJSON(data []byte) (v Verifier, err error) {
	t := BackupCodeVerifier{}
	if err = json.Unmarshal(data, &t); err != nil {
		return
	}
	v = t
	return
}
