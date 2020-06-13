package testutil

import (
	"testing"
)

func TestSetupTearDown(t *testing.T) {
	MigrationsDir = "../../db/migrations"
	FixturesDir = "../../db/fixtures"
	DBSetUp()
	defer DBTearDown()
}
