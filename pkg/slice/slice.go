package slice

import "reflect"

// Contains checks if a slice contains an element
func Contains(s interface{}, t interface{}) bool {
	return Index(s, t) != -1
}

// Index searches for x in a slice of ints using linear search and returns the index as specified by Search.
// It is used when the slice is small (e.g. smaller than 256 items) because linear search is faster in this case.
func Index(s interface{}, t interface{}) int {
	slice := convertSliceToInterface(s)
	for i, a := range slice {
		if a == t {
			return i
		}
	}
	return -1
}

// IsZeroBytes checks if the byte slice contains all zeros.
func IsZeroBytes(bytes []byte) bool {
    for _, v := range bytes {
        if v != 0 {
            return false
        }
    }
    return true
}

// convertSliceToInterface takes a slice passed in as an interface{}
// then converts the slice to a slice of interfaces
func convertSliceToInterface(s interface{}) (slice []interface{}) {
	v := reflect.ValueOf(s)
	if v.Kind() != reflect.Slice {
		return nil
	}

	length := v.Len()
	slice = make([]interface{}, length)
	for i := 0; i < length; i++ {
		slice[i] = v.Index(i).Interface()
	}

	return slice
}
