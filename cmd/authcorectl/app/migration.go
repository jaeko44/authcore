package app

import (
	"log"
	"net/url"

	"authcore.io/authcore/internal/config"
	secretutil "authcore.io/authcore/pkg/secret"

	"github.com/amacneil/dbmate/pkg/dbmate"
	"github.com/golang-migrate/migrate"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var db *dbmate.DB

// migrationCmd is a cobra command for running database migrations.
var migrationCmd = &cobra.Command{
	Use:   "migration",
	Short: "Database migration tools",

	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		initConfig()
		initDbmate()
	},
}

func initConfig() {
	config.InitDefaults()
	config.InitConfig()
}

func initDbmate() {
	databaseURLString, ok := viper.Get("database_url").(secretutil.String)
	if !ok {
		log.Fatalf("DATABASE_URL does not exist")
		return
	}
	databaseURL := databaseURLString.SecretString()
	u, err := url.Parse(databaseURL)
	if err != nil {
		log.Fatalf("DATABASE_URL is invalid: %v", err)
	}
	db = dbmate.New(u)
	db.MigrationsDir = viper.GetString("migration_dir")
	db.AutoDumpSchema = false
}

func createAndMigrate() {
	err := db.CreateAndMigrate()
	if err != nil {
		log.Fatalf("Migration fails: %v", err)
	}
}

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Creates the database (if necessary) and runs migrations",

	Run: func(cmd *cobra.Command, args []string) {
		createAndMigrate()
	},
}

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Rolls back the most recent migration",
	Run: func(cmd *cobra.Command, args []string) {
		err := db.Rollback()
		if err != nil {
			log.Fatalf("Rollback fails: %v", err)
		}
	},
}

func print(m *migrate.Migrate) {
	v, dirty, err := m.Version()
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	if dirty {
		log.Printf("%v (dirty)\n", v)
	} else {
		log.Println(v)
	}
}

func init() {
	migrationCmd.AddCommand(upCmd, downCmd)
}
