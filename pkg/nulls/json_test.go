package nulls

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewJSONNil(t *testing.T) {
	r := require.New(t)

	json := NewJSON(nil)

	r.False(json.Valid)
}

func TestNewJSONValidJSONButInvalidMap(t *testing.T) {
	r := require.New(t)

	json := NewJSON(`[{}]`)

	r.False(json.Valid)
}

func TestNewJSONEmpty(t *testing.T) {
	r := require.New(t)

	json := NewJSON(struct{}{})

	r.True(json.Valid)

	r.Equal(reflect.Map, reflect.TypeOf(json.Struct).Kind())
	r.Equal(0, len(json.Struct))
}

func TestNewJSONEmptyString(t *testing.T) {
	r := require.New(t)

	json := NewJSON(``)

	r.False(json.Valid)
}

func TestNewJSONEmptyObjectString(t *testing.T) {
	r := require.New(t)

	json := NewJSON(`{}`)

	r.True(json.Valid)
}

func TestNewJSONFromStruct(t *testing.T) {
	r := require.New(t)
	f := struct {
		Name string
	}{
		"FromStruct",
	}

	json := NewJSON(f)

	r.True(json.Valid)

	name, ok := json.Struct["Name"]
	r.True(ok)
	r.Equal(name, "FromStruct")
}

func TestNewJSONFromStructWithNil(t *testing.T) {
	r := require.New(t)
	f := struct {
		Name interface{}
	}{
		nil,
	}

	json := NewJSON(f)

	r.True(json.Valid)

	name, ok := json.Struct["Name"]
	r.True(ok)
	r.Equal(name, nil)
}

func TestNewJSONInString(t *testing.T) {
	r := require.New(t)

	json := NewJSON(`{"Name":"NotFromByte","Age":6,"Parents":["Gomez","Morticia"]}`)

	r.True(json.Valid)

	name, ok := json.Struct["Name"]
	r.True(ok)
	r.Equal(name, "NotFromByte")
	parents, ok := json.Struct["Parents"]
	parents1, _ := parents.([]interface{})
	r.Equal(parents1[0], "Gomez")
}

func TestNewJSONInStringNull(t *testing.T) {
	r := require.New(t)

	json := NewJSON(`{"Name":"NotFromByte","Age":null,"Parents":["Gomez","Morticia"]}`)

	r.True(json.Valid)

	age, ok := json.Struct["Age"]
	r.True(ok)
	r.Equal(age, nil)
}

func TestNewJSONInByteString(t *testing.T) {
	r := require.New(t)

	json := NewJSON([]byte(`{"Name":"NotFromByte","Age":6,"Parents":["Gomez","Morticia"]}`))

	r.True(json.Valid)

	name, ok := json.Struct["Name"]
	r.True(ok)
	r.Equal(name, "NotFromByte")
	parents, ok := json.Struct["Parents"]
	parents1, _ := parents.([]interface{})
	r.Equal(parents1[0], "Gomez")
}

func TestScan(t *testing.T) {
	r := require.New(t)
	json := NewJSON(struct{}{})

	json.Scan([]byte(`{"Name":"Scan","Age":6,"Parents":["Gomez","Morticia"]}`))

	r.True(json.Valid)
	name, ok := json.Struct["Name"]
	r.Equal(name, "Scan")
	r.True(ok)
}

func TestScanEmpty(t *testing.T) {
	r := require.New(t)
	json := NewJSON(``)

	json.Scan([]byte(``))

	r.Equal(json.Struct, map[string]interface{}(nil))
	r.Len(json.Struct, 0)
	r.False(json.Valid)
}
