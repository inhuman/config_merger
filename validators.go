package config_merger

import (
	"errors"
	"github.com/hashicorp/go-multierror"
	"reflect"
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

		//TODO: add checks for other tags

		column := field.Tag.Get("vault")

		if column != "" {
			for _, tag := range tags {
				if tag == column {
					return errors.New("duplicate vault tags found: " + string(tag))
				}
			}

			tags = append(tags, column)
		}
	}

	return nil
}

func (m *Merger) checkRequiredFields() error {

	t := reflect.TypeOf(m.TargetConfigStruct).Elem()
	v := reflect.ValueOf(m.TargetConfigStruct).Elem()

	var errAll *multierror.Error

	errAll = processRequiredTags(t, v, errAll, "")
	if errAll != nil {
		return errAll
	}

	return nil
}

func processRequiredTags(t reflect.Type, v reflect.Value, err *multierror.Error, parentConfig string) *multierror.Error {

	parentConfig += t.Name() + "."

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		if field.Type.Kind() == reflect.Struct {
			err = processRequiredTags(field.Type, value, err, parentConfig)
		}

		column := field.Tag.Get("required")

		if column == "true" {

			switch value.Kind() {
			case reflect.String:
				if value.String() == "" {
					newErr := errors.New("Required value " + parentConfig + field.Name + " is empty")
					err = multierror.Append(err, newErr)
				}

			case reflect.Int:
				if value.Int() == 0 {
					newErr := errors.New("Required value " + parentConfig + field.Name + " is 0 (can not be)")
					err = multierror.Append(err, newErr)
				}
			}
		}
	}

	return err
}
