package db

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type model struct {
	ID int32  `db:"id"`
	F  string `db:"f" fieldtag:"insert,update"`
	G  string `db:"g" fieldtag:"update"`
}

func TestModel(t *testing.T) {
	mm := &ModelMap{
		models: make(map[reflect.Type]*Model),
	}
	_ = mm.AddModel(new(model), "m")

	i := &model{F: "test"}
	ib, err := mm.InsertBuilder(i)
	assert.NoError(t, err)
	iq, ia := ib.Build()

	assert.Equal(t, "INSERT INTO m (f) VALUES (?)", iq)
	assert.Equal(t, []interface{}{"test"}, ia)

	sb, err := mm.SelectBuilder(i, 123)
	assert.NoError(t, err)
	sq, sa := sb.Build()

	assert.Equal(t, "SELECT m.id, m.f, m.g FROM m WHERE id = ?", sq)
	assert.Equal(t, []interface{}{int(123)}, sa)

	ub, err := mm.UpdateBuilder(i, 123)
	assert.NoError(t, err)
	uq, ua := ub.Build()

	assert.Equal(t, "UPDATE m SET f = ?, g = ? WHERE id = ?", uq)
	assert.Equal(t, []interface{}{"test", "", 123}, ua)
}
