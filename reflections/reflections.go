// Copyright (c) 2013 Théo Crevon
//
// See the file LICENSE for copying permission.

/*
Package reflections provides high level abstractions above the
reflect library.

Reflect library is very low-level and as can be quite complex when it comes to do simple things like accessing a structure field value, a field tag...

The purpose of reflections package is to make developers life easier when it comes to introspect structures at runtime.
It's API is freely inspired from python language (getattr, setattr, hasattr...) and provides a simplified access to structure fields and tags.
*/

/*
This file is based on https://github.com/oleiade/reflections. I added some functions.
*/
package reflections

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// GetField returns the value of the provided obj field. obj can whether
// be a structure or pointer to structure.
func GetField(obj interface{}, name string) (interface{}, error) {
	if !hasValidType(obj, []reflect.Kind{reflect.Struct, reflect.Ptr}) {
		return nil, errors.New("getField can not use GetField on a non-struct interface")
	}

	objValue := reflectValue(obj)
	field := objValue.FieldByName(name)
	if !field.IsValid() {
		return nil, fmt.Errorf("getField no such field: %s in obj", name)
	}

	return field.Interface(), nil
}

// HasField checks if the provided field name is part of a struct. obj can whether
// be a structure or pointer to structure.
func HasField(obj interface{}, name string) (bool, error) {
	if !hasValidType(obj, []reflect.Kind{reflect.Struct, reflect.Ptr}) {
		return false, errors.New("setField can not use GetField on a non-struct interface")
	}

	objValue := reflectValue(obj)
	objType := objValue.Type()
	field, ok := objType.FieldByName(name)
	if !ok || !isExportableField(field) {
		return false, nil
	}

	return true, nil
}

func reflectValue(obj interface{}) reflect.Value {
	var val reflect.Value

	if reflect.TypeOf(obj).Kind() == reflect.Ptr {
		val = reflect.ValueOf(obj).Elem()
	} else {
		val = reflect.ValueOf(obj)
	}

	return val
}

func isExportableField(field reflect.StructField) bool {
	// PkgPath is empty for exported fields.
	return field.PkgPath == ""
}

func hasValidType(obj interface{}, types []reflect.Kind) bool {
	for _, t := range types {
		if reflect.TypeOf(obj).Kind() == t {
			return true
		}
	}

	return false
}

func getFieldNameByJsonTagFromType(t reflect.Type, jsonTag string) string {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	for i := 0; i < t.NumField(); i++ {
		fieldType := t.Field(i).Type
		if fieldType.Kind() == reflect.Struct || (fieldType.Kind() == reflect.Ptr && fieldType.Elem().Kind() == reflect.Struct) {
			if k := getFieldNameByJsonTagFromType(fieldType, jsonTag); k != "" {
				return k
			}
		}

		jsonStr := t.Field(i).Tag.Get("json")
		for _, v := range strings.Split(jsonStr, ",") {
			if v == jsonTag {
				return t.Field(i).Name
			}
		}
	}

	return ""
}

func GetFieldNameByJsonTag(obj interface{}, jsonTag string) (string, error) {
	t := reflect.TypeOf(obj)
	fieldName := getFieldNameByJsonTagFromType(t, jsonTag)
	if fieldName != "" {
		return fieldName, nil
	}

	return "", fmt.Errorf("GetFieldByJsonTag no such tag: %s in obj: %v", jsonTag, obj)
}
