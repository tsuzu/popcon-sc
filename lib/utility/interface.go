package utility

import "reflect"

func dereferenceIfPtr(value interface{}) interface{} {
	if reflect.TypeOf(value).Kind() == reflect.Ptr {
		return reflect.ValueOf(value).Elem().Interface()
	}

	return value
}
