package goatquery

import (
	"reflect"
	"strings"
)

func BuildPagedResponse[T any](res []T, query Query, totalCount *int64) PagedResponse[map[string]interface{}] {
	result := make([]map[string]interface{}, len(res))

	selectedProperties := strings.Split(strings.TrimSpace(query.Select), ",")

	for i, obj := range res {
		newObj := make(map[string]interface{})
		t := reflect.TypeOf(obj)
		v := reflect.ValueOf(obj)

		if query.Select != "" {
			// map over selected properties
			for _, p := range selectedProperties {
				property := strings.TrimSpace(p)
				field, _ := v.Type().FieldByName(property)
				name := field.Tag.Get("json")

				if name != "" && name != "-" {
					// '-' in the json tag means to not return that property
					newObj[name] = v.FieldByName(property).String()
				}
			}
		} else {
			// map over every property
			for i := 0; i < t.NumField(); i++ {
				field := t.Field(i)
				value := v.Field(i).Interface()

				name := field.Tag.Get("json")

				if name != "" && name != "-" {
					// '-' in the json tag means to not return that property
					newObj[name] = value
				}
			}
		}

		result[i] = newObj
	}

	return PagedResponse[map[string]interface{}]{Value: result, Count: totalCount}
}
