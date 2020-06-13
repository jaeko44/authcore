package db

import (
	"reflect"

	"github.com/huandu/go-sqlbuilder"
	"github.com/jmoiron/sqlx/reflectx"
	"github.com/pkg/errors"
)

const (
	insertTag = "insert"
	updateTag = "update"
)

// ModelMap is a simple database model mapper based on sqlbuilder.
type ModelMap struct {
	models map[reflect.Type]*Model
	mapper *reflectx.Mapper
}

// AddModel registers the given interface type as a model for SQL builder.
func (t *ModelMap) AddModel(i interface{}, tableName string, keys ...string) *Model {
	ty := reflect.TypeOf(i)
	if tableName == "" {
		tableName = ty.Name()
	}
	var key string
	if len(keys) == 0 {
		key = "id"
	} else {
		key = keys[0]
	}

	// check if we have a table for this type already
	// if so, update the name and return the existing pointer
	m, ok := t.models[ty]
	if ok {
		m.TableName = tableName
		return m
	}

	m = &Model{
		TableName: tableName,
		Key:       key,
		Builder:   sqlbuilder.NewStruct(i),
		Type:      ty,
	}

	t.models[ty] = m
	return m
}

// Model returns a *Model corresponding to the given type.
func (t *ModelMap) Model(i interface{}) (*Model, error) {
	ty := reflect.TypeOf(i)
	m, ok := t.models[ty]
	if ok {
		return m, nil
	}
	return nil, errors.Errorf("type %v is not registered", ty.Name())
}

// InsertBuilder creates a new `InsertBuilder` for fields with fieldtag "insert".
func (t *ModelMap) InsertBuilder(i interface{}) (*sqlbuilder.InsertBuilder, error) {
	m, err := t.Model(i)
	if err != nil {
		return nil, err
	}

	return m.Builder.InsertIntoForTag(m.TableName, insertTag, i), nil
}

// SelectBuilder creates a new `SelectBuilder` selecting the given primary key.
func (t *ModelMap) SelectBuilder(i interface{}, key interface{}) (*sqlbuilder.SelectBuilder, error) {
	m, err := t.Model(i)
	if err != nil {
		return nil, err
	}

	sb := m.Builder.SelectFrom(m.TableName)
	sb.Where(sb.E(m.Key, key))
	return sb, nil
}

// SelectAllBuilder creates a new `SelectBuilder` for all rows.
func (t *ModelMap) SelectAllBuilder(i interface{}) (*sqlbuilder.SelectBuilder, error) {
	m, err := t.Model(i)
	if err != nil {
		return nil, err
	}

	sb := m.Builder.SelectFrom(m.TableName)
	return sb, nil
}

// UpdateBuilder creates a new `SelectBuilder` for fields with fieldtag "update".
func (t *ModelMap) UpdateBuilder(i interface{}, key interface{}) (*sqlbuilder.UpdateBuilder, error) {
	m, err := t.Model(i)
	if err != nil {
		return nil, err
	}

	ub := m.Builder.UpdateForTag(m.TableName, updateTag, i)
	ub.Where(ub.E(m.Key, key))
	return ub, nil
}

// Model represents a mapping between a Go struct and a *sqlbuilder.Struct
type Model struct {
	TableName string
	Key       string
	Builder   *sqlbuilder.Struct
	Type      reflect.Type
}
