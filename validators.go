package config_merger

import (
	"reflect"
	"errors"
)

type validator func(s interface{}) error

var validators []validator

func init() {
	validators = append(validators, ValidateVaultDuplicates)
}

func validateStruct(s interface{}) error {

	for _, f := range validators {
		err := f(s)
		if err != nil {
			return err
		}
	}
	return nil
}

func ValidateVaultDuplicates(s interface{}) error {
	var tags []string
	return isDuplicateTags(reflect.TypeOf(s).Elem(), tags)
}

func isDuplicateTags(t reflect.Type, tags []string) error {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		if field.Type.Kind() == reflect.Struct {

			err := isDuplicateTags(field.Type, tags)
			if err != nil {
				return err
			}
			continue
		}

		column := field.Tag.Get("vault")

		for _, tag := range tags {
			if tag == column {
				return errors.New("duplicate vault tags found: " + string(tag))
			}
		}

		tags = append(tags, column)
	}

	return nil
}
