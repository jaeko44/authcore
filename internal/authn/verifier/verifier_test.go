package verifier

import (
	"os"
	"testing"

	"authcore.io/authcore/internal/testutil"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	testutil.MigrationsDir = "../../../db/migrations"
	testutil.FixturesDir = "../../../db/fixtures"
	testutil.DBSetUp()
	code := m.Run()
	testutil.DBTearDown()
	os.Exit(code)
}

func TestUnmarshalPasswordVerifier(t *testing.T) {
	f := NewFactory()

	data := `{
		"method": "spake2plus",
		"salt": "/Jb4pAuatq5rrwdNRGRqW+PhlqzNR1pYtp1N5YWEn7s=",
		"w0": "MK43qvO3DoflMLl1AlZPJA==",
		"l": "gGQMM78ixkqSUTxq3LvyLuURe1bXQSUbqCPsEHfg65M="
	}`

	verifier, err := f.Unmarshal([]byte(data))
	if !assert.NoError(t, err) {
		return
	}
	_, ok := verifier.(SPAKE2PlusVerifier)
	assert.True(t, ok)

	// Invalid method
	data2 := `{
		"method": "invalid",
		"salt": "/Jb4pAuatq5rrwdNRGRqW+PhlqzNR1pYtp1N5YWEn7s=",
		"w0": "MK43qvO3DoflMLl1AlZPJA==",
		"l": "gGQMM78ixkqSUTxq3LvyLuURe1bXQSUbqCPsEHfg65M="
	}`

	_, err2 := f.Unmarshal([]byte(data2))
	if !assert.Error(t, err2) {
		return
	}
}
