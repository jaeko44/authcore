package db

import (
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"time"

	"authcore.io/authcore/pkg/secret"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

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
