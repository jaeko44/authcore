// Package testutil contains helpers for testing the core package
package testutil

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"os"
	"time"

	"authcore.io/authcore/pkg/secret"

	"github.com/amacneil/dbmate/pkg/dbmate"
	"github.com/spf13/viper"
	"github.com/xo/dburl"
	testfixtures "gopkg.in/testfixtures.v2"
)

// MigrationsDir is the path to migration scripts.
var MigrationsDir = "../../db/migrations"

// FixturesDir is the path to fixtures data.
var FixturesDir = "../../db/fixtures"

// DBSetUp initializes a database for tests.
func DBSetUp() {
	configureDB()
	db := getDbmate()
	db.Drop() // ignore error
	err := db.CreateAndMigrate()
	if err != nil {
		cwd, _ := os.Getwd()
		log.Printf("Current directory: %v", cwd)
		log.Fatalf("Database migration failed: %v", err)
	}
	FixturesSetUp()
}

// DBTearDown destroy the database for tests.
func DBTearDown() {
	db := getDbmate()
	err := db.Drop()
	if err != nil {
		log.Fatalf("Drop database failed: %v", err)
	}
}

func configureDB() {
	rand.Seed(time.Now().UnixNano())
	databaseURL := os.Getenv("TEST_DATABASE_URL")

	u, err := url.Parse(databaseURL)
	if err != nil {
		log.Fatalf("required environment TEST_DATABASE_URL is invalid: %v", err)
	}
	u.Path = fmt.Sprintf("%v_%v", u.Path, rand.Uint32())
	databaseURL = u.String()

	log.Printf("Setting DATABASE_URL = %v", databaseURL)
	os.Setenv("DATABASE_URL", databaseURL)
	viper.Set("database_url", secret.NewString(databaseURL))
}

func getDbmate() *dbmate.DB {
	u, err := url.Parse(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("DATABASE_URL is invalid: %v", err)
	}
	db := dbmate.New(u)
	db.MigrationsDir = MigrationsDir
	db.AutoDumpSchema = false

	return db
}

// FixturesSetUp setup a fixture for test
func FixturesSetUp() {
	databaseURL := os.Getenv("DATABASE_URL")
	u, err := dburl.Parse(databaseURL)
	if err != nil {
		log.Fatalf("cannot parse DATABASE_URL: %v", err)
	}
	viper.Set("database_url", secret.NewString(databaseURL))

	db, err := sql.Open(u.Driver, u.DSN)
	if err != nil {
		log.Fatalf("failed to open database for fixtures: %v", err)
	}
	fixtures, err := testfixtures.NewFolder(db, &testfixtures.MySQL{}, FixturesDir)
	if err != nil {
		log.Fatalf("failed to prepare fixtures: %v", err)
	}
	if err := fixtures.Load(); err != nil {
		log.Fatalf("failed to load fixtures: %v", err)
	}
	db.Close()
}
