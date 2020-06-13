package paging

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/xo/dburl"

	// Loads the MySQL driver
	_ "github.com/go-sql-driver/mysql"
)

func TestPrepare(t *testing.T) {
	var options PageOptions
	var query string
	var args []interface{}
	var err error

	options = PageOptions{
		SortColumn:   "updated_at",
		UniqueColumn: "id",
		Limit:        50,
	}
	query, args, err = prepare(options, nil, "SELECT * FROM tx")
	if assert.NoError(t, err) {
		assert.Equal(t, "SELECT * FROM (SELECT * FROM tx) AS t ORDER BY `updated_at` ASC, `id` ASC LIMIT ?", query)
		assert.Equal(t, []interface{}{uint(51)}, args)
	}

	options = PageOptions{
		SortColumn:    "updated_at",
		UniqueColumn:  "id",
		SortDirection: Asc,
		Limit:         50,
	}
	query, args, err = prepare(options, nil, "SELECT * FROM tx WHERE user_id=?", 10)
	if assert.NoError(t, err) {
		assert.Equal(t, "SELECT * FROM (SELECT * FROM tx WHERE user_id=?) AS t ORDER BY `updated_at` ASC, `id` ASC LIMIT ?", query)
		assert.Equal(t, []interface{}{10, uint(51)}, args)
	}

	options = PageOptions{
		SortColumn:    "updated_at",
		UniqueColumn:  "id",
		SortDirection: Desc,
		Limit:         100,
	}
	query, args, err = prepare(options, nil, "SELECT * FROM tx WHERE user_id=?", 10)
	if assert.NoError(t, err) {
		assert.Equal(t, "SELECT * FROM (SELECT * FROM tx WHERE user_id=?) AS t ORDER BY `updated_at` DESC, `id` DESC LIMIT ?", query)
		assert.Equal(t, []interface{}{10, uint(101)}, args)
	}

	options = PageOptions{
		SortColumn:    "updated_at",
		UniqueColumn:  "id",
		SortDirection: Desc,
		Limit:         100,
	}
	cursor := PageCursor{
		Direction: Next,
		Cursor:    []interface{}{"test", float64(12423434)},
	}
	query, args, err = prepare(options, &cursor, "SELECT * FROM tx WHERE user_id=?", 10)
	if assert.NoError(t, err) {
		assert.Equal(t, "SELECT * FROM (SELECT * FROM tx WHERE user_id=?) AS t WHERE (`updated_at`, `id`) < (?, ?) ORDER BY `updated_at` DESC, `id` DESC LIMIT ?", query)
		assert.Equal(t, []interface{}{10, "test", float64(12423434), uint(101)}, args)
	}

	// Invalid columns
	options = PageOptions{
		SortColumn:   "id\n`xxx",
		UniqueColumn: "id",
		Limit:        50,
	}
	query, args, err = prepare(options, nil, "SELECT * FROM tx WHERE user_id=?", 10)
	if assert.Error(t, err) {
		assert.Equal(t, "options.SortColumn contains invalid characters", err.Error())
	}

	options = PageOptions{
		SortColumn:   "id",
		UniqueColumn: "id\n`xxx",
		Limit:        50,
	}
	query, args, err = prepare(options, nil, "SELECT * FROM tx WHERE user_id=?", 10)
	if assert.Error(t, err) {
		assert.Equal(t, "options.UniqueColumn contains invalid characters", err.Error())
	}
}

func TestSelectContext(t *testing.T) {
	db, tearDown := dbopen(os.Getenv("TEST_DATABASE_URL"))
	defer tearDown()
	db.MustExec("CREATE TABLE paging_test (`a` int PRIMARY KEY, `b` int NOT NULL)")
	for i := 1; i < 9; i++ {
		db.MustExec("INSERT INTO paging_test (a, b) VALUE (?, ?)", i, (10-i)/3)
	}
	type PagingTest struct {
		A int `db:"a"`
		B int `db:"b"`
	}

	expectedPages := [][]int{
		[]int{8, 5, 6},
		[]int{7, 2, 3},
		[]int{4, 1},
	}

	// Page 1
	pageOptions := PageOptions{
		SortColumn:   "b",
		UniqueColumn: "a",
		Limit:        3,
	}
	var results []PagingTest
	page, err := SelectContext(context.Background(), db, pageOptions, &results, "SELECT * FROM paging_test")
	if assert.NoError(t, err) {
		assert.Len(t, results, len(expectedPages[0]))
		for i, row := range results {
			assert.Equal(t, expectedPages[0][i], row.A)
		}
		assert.Empty(t, page.PreviousPageToken)
		assert.NotEmpty(t, page.NextPageToken)
		assert.Equal(t, uint(0), page.FoundRows)
		assert.Equal(t, 1, db.Stats().OpenConnections)
	}

	// Next of page 1
	pageOptions = PageOptions{
		SortColumn:   "b",
		UniqueColumn: "a",
		Limit:        3,
		PageToken:    page.NextPageToken,
	}
	results = []PagingTest{}
	page, err = SelectContext(context.Background(), db, pageOptions, &results, "SELECT * FROM paging_test")
	if assert.NoError(t, err) {
		assert.Len(t, results, len(expectedPages[1]))
		for i, row := range results {
			assert.Equal(t, expectedPages[1][i], row.A)
		}
		assert.NotEmpty(t, page.PreviousPageToken)
		assert.NotEmpty(t, page.NextPageToken)
		assert.Equal(t, uint(0), page.FoundRows)
		assert.Equal(t, 1, db.Stats().OpenConnections)
	}

	// Next of page 2
	pageOptions = PageOptions{
		SortColumn:   "b",
		UniqueColumn: "a",
		Limit:        3,
		PageToken:    page.NextPageToken,
	}
	results = []PagingTest{}
	page, err = SelectContext(context.Background(), db, pageOptions, &results, "SELECT * FROM paging_test")
	if assert.NoError(t, err) {
		assert.Len(t, results, len(expectedPages[2]))
		for i, row := range results {
			assert.Equal(t, expectedPages[2][i], row.A)
		}
		assert.NotEmpty(t, page.PreviousPageToken)
		assert.Empty(t, page.NextPageToken)
		assert.Equal(t, uint(0), page.FoundRows)
		assert.Equal(t, 1, db.Stats().OpenConnections)
	}

	// Previous of page 3
	pageOptions = PageOptions{
		SortColumn:   "b",
		UniqueColumn: "a",
		Limit:        3,
		PageToken:    page.PreviousPageToken,
	}
	results = []PagingTest{}
	page, err = SelectContext(context.Background(), db, pageOptions, &results, "SELECT * FROM paging_test")
	if assert.NoError(t, err) {
		assert.Len(t, results, len(expectedPages[1]))
		for i, row := range results {
			assert.Equal(t, expectedPages[1][i], row.A)
		}
		assert.NotEmpty(t, page.PreviousPageToken)
		assert.NotEmpty(t, page.NextPageToken)
		assert.Equal(t, uint(0), page.FoundRows)
		assert.Equal(t, 1, db.Stats().OpenConnections)
	}

	// Previous of page 2
	pageOptions = PageOptions{
		SortColumn:   "b",
		UniqueColumn: "a",
		Limit:        3,
		PageToken:    page.PreviousPageToken,
	}
	results = []PagingTest{}
	page, err = SelectContext(context.Background(), db, pageOptions, &results, "SELECT * FROM paging_test")
	if assert.NoError(t, err) {
		assert.Len(t, results, len(expectedPages[0]))
		for i, row := range results {
			assert.Equal(t, expectedPages[0][i], row.A)
		}
		assert.Empty(t, page.PreviousPageToken)
		assert.NotEmpty(t, page.NextPageToken)
		assert.Equal(t, uint(0), page.FoundRows)
		assert.Equal(t, 1, db.Stats().OpenConnections)
	}

	// Count found rows
	pageOptions = PageOptions{
		CountFoundRows: true,
		SortColumn:     "b",
		UniqueColumn:   "a",
		Limit:          3,
	}
	results = []PagingTest{}
	page, err = SelectContext(context.Background(), db, pageOptions, &results, "SELECT * FROM paging_test")
	if assert.NoError(t, err) {
		assert.Len(t, results, 3)
		for i, row := range results {
			assert.Equal(t, expectedPages[0][i], row.A)
		}
		assert.Empty(t, page.PreviousPageToken)
		assert.NotEmpty(t, page.NextPageToken)
		assert.Equal(t, uint(8), page.FoundRows)
		// Check there is no extra connection through the query process
		assert.Equal(t, 1, db.Stats().OpenConnections)
	}
}

func dbopen(databaseURL string) (*sqlx.DB, func()) {
	u, err := url.Parse(databaseURL)
	if err != nil {
		log.Fatalf("required environment TEST_DATABASE_URL is invalid: %v", err)
	}
	u.Path = fmt.Sprintf("%v_%v", u.Path, "paging")
	dbname := databaseName(u)

	if !IsValidColumn(dbname) {
		log.Fatalf("required environment TEST_DATABASE_URL is invalid: %v", err)
	}

	dbu, err := dburl.Parse(u.String())
	if err != nil {
		log.Fatalf("required environment TEST_DATABASE_URL is invalid: %v", err)
	}
	db := sqlx.MustOpen(dbu.Driver, dbu.DSN)

	u.Path = "/"
	rootu, err := dburl.Parse(u.String())
	if err != nil {
		log.Fatalf("required environment TEST_DATABASE_URL is invalid: %v", err)
	}
	rootdb := sqlx.MustOpen(rootu.Driver, rootu.DSN)
	rootdb.Exec("DROP DATABASE " + dbname)
	rootdb.MustExec("CREATE DATABASE " + dbname)

	return db, func() {
		// rootdb.Exec("DROP DATABASE " + dbname)
		db.Close()
		rootdb.Close()
	}
}

// databaseName returns the database name from a URL
func databaseName(u *url.URL) string {
	name := u.Path
	if len(name) > 0 && name[:1] == "/" {
		name = name[1:]
	}

	return name
}

func TestIsValidColumn(t *testing.T) {
	assert.True(t, IsValidColumn("id"))
	assert.True(t, IsValidColumn("id$"))
	assert.True(t, IsValidColumn("user_name"))
	assert.False(t, IsValidColumn("id`"))
	assert.False(t, IsValidColumn("'id"))
	assert.False(t, IsValidColumn("user`name"))
	assert.False(t, IsValidColumn("id\n`xxx"))
	assert.False(t, IsValidColumn("用戶名"))
}
