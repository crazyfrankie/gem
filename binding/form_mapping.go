package binding

import (
	"errors"
	"reflect"
)

var ( // ErrConvertMapStringSlice can not convert to map[string][]string
	ErrConvertMapStringSlice = errors.New("can not convert to map slices of strings")

	// ErrConvertToMapString can not convert to map[string]string
	ErrConvertToMapString = errors.New("can not convert to map of strings")
)

func mapForm(ptr any, form map[string][]string) error {
	return mapFormByTag(ptr, form, "form")
}

func mapFormByTag(ptr any, form map[string][]string, tag string) error {
	ptrVal := reflect.ValueOf(ptr)
	var pointed any
	if ptrVal.Kind() == reflect.Ptr {
		ptrVal = ptrVal.Elem()
		pointed = ptrVal.Interface()
	}

	if ptrVal.Kind() == reflect.Map && ptrVal.Type().Key().Kind() == reflect.String {
		if pointed != nil {
			ptr = pointed
		}
		return setMapForm(ptr, form)
	}

	return mappingByPtr(ptr, form)
}

func setMapForm(ptr any, form map[string][]string) error {
	el := reflect.TypeOf(ptr).Elem()

	if el.Kind() == reflect.Slice {
		ptrMap, ok := ptr.(map[string][]string)
		if !ok {
			return ErrConvertMapStringSlice
		}
		for k, v := range form {
			ptrMap[k] = v
		}
	}

	ptrMap, ok := ptr.(map[string]string)
	if !ok {
		return ErrConvertToMapString
	}
	for k, v := range form {
		ptrMap[k] = v[len(v)-1] // pick last
	}

	return nil
}

func mappingByPtr(ptr any, form map[string][]string) error {
	return nil
}
