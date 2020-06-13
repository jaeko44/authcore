package paging

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsePageCursor(t *testing.T) {
	cursor, err := parsePageCursor("eyJkIjowLCJ2IjpbInRlc3QiLDEyNDIzNDM0XX0")

	assert.NoError(t, err)
	assert.Equal(t, Next, cursor.Direction)
	assert.Equal(t, []interface{}{"test", float64(12423434)}, cursor.Cursor)
}

func TestNewPageToken(t *testing.T) {
	token := newPageToken(Next, []interface{}{"test", 12423434})
	assert.Equal(t, "eyJkIjowLCJ2IjpbInRlc3QiLDEyNDIzNDM0XX0", string(token))
}

func TestParseSortBy(t *testing.T) {
	col, dir, err := ParseSortBy("foo")
	assert.Equal(t, "foo", col)
	assert.Equal(t, Asc, dir)
	assert.NoError(t, err)

	col, dir, err = ParseSortBy("FOO$bar")
	assert.Equal(t, "FOO$bar", col)
	assert.Equal(t, Asc, dir)
	assert.NoError(t, err)

	col, dir, err = ParseSortBy("    FOO$bar     asc  ")
	assert.Equal(t, "FOO$bar", col)
	assert.Equal(t, Asc, dir)
	assert.NoError(t, err)

	col, dir, err = ParseSortBy("    FOO$bar     DEsc  ")
	assert.Equal(t, "FOO$bar", col)
	assert.Equal(t, Desc, dir)
	assert.NoError(t, err)

	col, dir, err = ParseSortBy("foo,bar")
	assert.Error(t, err)

	_, _, err = ParseSortBy("foo-bar")
	assert.Error(t, err)
}
