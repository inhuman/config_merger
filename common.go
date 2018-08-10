package config_merger

import (
	"reflect"
	"encoding/json"
)

func processTags(t reflect.Type, v reflect.Value, b []byte, callback func(i interface{}) (string, error)) error {

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		if field.Type.Kind() == reflect.Struct {
			err := processTags(field.Type, value, b, callback)
			if err != nil {
				return err
			}
			continue
		}

		column := field.Tag.Get("vault")

		if column != "" {

			var bm map[string]string
			bm = make(map[string]string)
			err := json.Unmarshal(b, &bm)
			if err != nil {
				return nil
			}

			//TODO: add int and float types, just in case
			value.SetString(bm[column])
		}
	}

	return nil
}