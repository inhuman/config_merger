package config_merger

import (
	"github.com/hashicorp/go-multierror"
	"reflect"
	"strconv"
)

func (m *Merger) setDefaults() error {

	t := reflect.TypeOf(m.TargetConfigStruct).Elem()
	v := reflect.ValueOf(m.TargetConfigStruct).Elem()

	var errAll *multierror.Error

	errAll = processDefaultsTags(t, v, errAll)
	if errAll != nil {
		return errAll
	}

	return nil

}

func processDefaultsTags(t reflect.Type, v reflect.Value, err *multierror.Error) *multierror.Error {

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		if field.Type.Kind() == reflect.Struct {
			err = processDefaultsTags(field.Type, value, err)
		}

		column := field.Tag.Get("default")

		if column != "" {
			switch value.Kind() {
			case reflect.String:
				if value.String() == "" {
					value.SetString(column)
				}

			case reflect.Int:
				if value.Int() == 0 {
					i, newErr := strconv.ParseInt(column, 10, 64)
					if newErr != nil {
						err = multierror.Append(err, newErr)
					} else {
						value.SetInt(i)
					}
				}
			}
		}
	}

	return err
}
