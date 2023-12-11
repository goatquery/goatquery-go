package goatquery

import (
	"encoding/json"
	"reflect"
	"strings"
)

func BuildPagedResponse[T any](res []T, query Query, totalCount *int64) PagedResponse[map[string]interface{}] {
	result := make([]map[string]interface{}, len(res))

	selectedProperties := strings.Split(strings.TrimSpace(query.Select), ",")

	if query.Select == "" {
		bytes, _ := json.Marshal(res)

		if err := json.Unmarshal(bytes, &result); err != nil {
			return PagedResponse[map[string]interface{}]{Value: result, Count: totalCount}
		}

		return PagedResponse[map[string]interface{}]{Value: result, Count: totalCount}
	}

	for i, obj := range res {
		newObj := make(map[string]interface{})
		v := reflect.ValueOf(obj)

		// map over selected properties
		for _, p := range selectedProperties {
			property := strings.TrimSpace(p)
			field, _ := v.Type().FieldByNameFunc(func(p string) bool {
				return strings.EqualFold(property, p)
			})
			name := field.Tag.Get("json")

			// '-' in the json tag means to not return that property
			if name != "" && name != "-" {
				newObj[name] = v.FieldByNameFunc(func(p string) bool {
					return strings.EqualFold(property, p)
				})
			}
		}

		result[i] = newObj
	}

	return PagedResponse[map[string]interface{}]{Value: result, Count: totalCount}
}
