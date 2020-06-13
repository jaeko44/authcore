package db

import (
	"bytes"
	"time"

	"authcore.io/authcore/pkg/nulls"
)

// NullableString converts string to nulls.String while treating empty string as null.
func NullableString(s string) nulls.String {
	if s == "" {
		return nulls.String{}
	}
	return nulls.NewString(s)
}

// NullableTime converts time.Time to nulls.Time while treating zero time as null.
func NullableTime(t time.Time) nulls.Time {
	if t.IsZero() {
		return nulls.Time{}
	}
	return nulls.NewTime(t)
}

// NullableInt64 converts int64 to nulls.Int64 while treating empty int as null.
func NullableInt64(value interface{}) nulls.Int64 {
	i, ok := value.(int64)
	if !ok {
		return nulls.Int64{}
	}
	return nulls.NewInt64(i)
}

// NullableByteSlice converts string to nulls.ByteSlice while treating empty byte slice as null.
func NullableByteSlice(b []byte) nulls.ByteSlice {
	if bytes.Equal(b, []byte{}) {
		return nulls.ByteSlice{}
	}
	return nulls.NewByteSlice(b)
}
