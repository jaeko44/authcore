package db

import (
	"context"
	"database/sql"
	"reflect"
	"regexp"
	"strconv"
	"time"

	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/pkg/log"
	"authcore.io/authcore/internal/validator"
	"authcore.io/authcore/pkg/secret"

	"github.com/gchaincl/sqlhooks"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/xo/dburl"

	// Loads the MySQL driver
	"github.com/go-sql-driver/mysql"
)

// DB is a wrapper of *sqlx.DB that adds logging and contextual transaction.
type DB struct {
	*sqlx.DB
	ModelMap *ModelMap
}

// NewDBFromConfig returns a *DB according to config values.
func NewDBFromConfig() *DB {
	databaseURL := viper.Get("database_url").(secret.String).SecretString()
	if databaseURL == "" {
		logrus.Fatalf("missing DATABASE_URL configuration value")
	}
	db, err := dbopen(databaseURL)
	if err != nil {
		logrus.Fatal(err.Error())
	}

	err = db.Ping()
	if err != nil {
		logrus.Fatal(err.Error())
	}

	return &DB{
		DB: db,
		ModelMap: &ModelMap{
			models: make(map[reflect.Type]*Model),
			mapper: db.Mapper,
		},
	}
}

// RunInTransaction runs a function in a transaction. If function returns an error transaction is
// rollbacked, otherwise transaction is committed.
func (db *DB) RunInTransaction(ctx context.Context, fn func(*sqlx.Tx) error) error {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	err = fn(tx)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

// AddModel registers the given interface type as a model for SQL builder.
func (db *DB) AddModel(i interface{}, tableName string, key ...string) *Model {
	return db.ModelMap.AddModel(i, tableName, key...)
}

// InsertModel inserts the model and refresh the model.
func (db *DB) InsertModel(ctx context.Context, i interface{}) error {
	if err := validator.Validate.Struct(i); err != nil {
		return errors.WithValidateError(err)
	}

	ib, err := db.ModelMap.InsertBuilder(i)
	if err != nil {
		return err
	}

	iq, ia := ib.Build()
	result, err := db.ExecContext(ctx, iq, ia...)
	if err != nil {
		return errors.WithSQLError(err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return errors.WithSQLError(err)
	}

	sb, err := db.ModelMap.SelectBuilder(i, id)
	if err != nil {
		return err
	}
	sq, sa := sb.Build()
	err = db.GetContext(ctx, i, sq, sa...)
	if err != nil {
		return errors.WithSQLError(err)
	}
	return nil
}

// SelectModel selects the model by primary key.
func (db *DB) SelectModel(ctx context.Context, i interface{}) error {
	m, err := db.ModelMap.Model(i)
	if err != nil {
		return err
	}
	mapper := db.DB.Mapper
	f := mapper.FieldByName(reflect.ValueOf(i), m.Key)

	sb, err := db.ModelMap.SelectBuilder(i, f.Interface())
	if err != nil {
		return err
	}
	sq, sa := sb.Build()
	err = db.GetContext(ctx, i, sq, sa...)
	if err != nil {
		return errors.WithSQLError(err)
	}
	return nil
}

// UpdateModel updates the model by primary key.
func (db *DB) UpdateModel(ctx context.Context, i interface{}) (sql.Result, error) {
	if err := validator.Validate.Struct(i); err != nil {
		return nil, errors.WithValidateError(err)
	}

	m, err := db.ModelMap.Model(i)
	if err != nil {
		return nil, err
	}
	mapper := db.DB.Mapper
	f := mapper.FieldByName(reflect.ValueOf(i), m.Key)

	ub, err := db.ModelMap.UpdateBuilder(i, f.Interface())
	if err != nil {
		return nil, err
	}
	uq, ua := ub.Build()
	result, err := db.ExecContext(ctx, uq, ua...)
	if err != nil {
		return nil, errors.WithSQLError(err)
	}
	return result, nil
}

func dbopen(databaseURL string) (*sqlx.DB, error) {
	u, err := dburl.Parse(databaseURL)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "cannot parse DATABASE_URL")
	}

	db, err := sqlx.Open(u.Driver+"+hooks", u.DSN)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "cannot open database")
	}

	return db, nil
}

func init() {
	sql.Register("mysql+hooks", sqlhooks.Wrap(&mysql.MySQLDriver{}, &hooks{}))
}

type startTimeKey struct{}

var spaceRe = regexp.MustCompile(`\s+`)

type hooks struct{}

func (h *hooks) Before(ctx context.Context, query string, args ...interface{}) (context.Context, error) {
	return context.WithValue(ctx, startTimeKey{}, time.Now()), nil
}

func (h *hooks) After(ctx context.Context, query string, args ...interface{}) (context.Context, error) {
	begin := ctx.Value(startTimeKey{}).(time.Time)
	query = spaceRe.ReplaceAllString(query, " ")
	l := time.Since(begin)
	log.GetLogger(ctx).WithFields(logrus.Fields{
		"query":         query,
		"latency":       strconv.FormatInt(int64(l), 10),
		"latency_human": l.String(),
	}).Info("SQL")
	return ctx, nil
}
