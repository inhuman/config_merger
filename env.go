package config_merger

import (
	"os"
	"reflect"
	"strconv"
	"sync"
)

type EnvSource struct {
	SourceModel
	Variables []string
}

func (s *EnvSource) Load() error {

	t := reflect.TypeOf(s.TargetStruct).Elem()
	v := reflect.ValueOf(s.TargetStruct).Elem()

	err := s.processEnvTags(t, v)
	if err != nil {
		return err
	}

	return nil
}

func (s *EnvSource) processEnvTags(t reflect.Type, v reflect.Value) error {

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		if !v.IsValid() {
			continue
		}

		value := v.Field(i)

		if field.Type.Kind() == reflect.Struct {
			if err := s.processEnvTags(field.Type, value); err != nil {
				return err
			}
		}

		if field.Type.Kind() == reflect.Ptr {
			field.Type = field.Type.Elem()
			value = value.Elem()

			if err := s.processEnvTags(field.Type, value); err != nil {
				return err
			}
		}

		column := GetTagContents(s, "env", field)

		if (column != "") && (StringInSlice(column, s.Variables)) {
			v := os.Getenv(column)

			if v != "" {

				v = trimQuotes(v)

				switch value.Kind() {
				case reflect.String:
					value.SetString(v)

				case reflect.Int:
					i, err := strconv.ParseInt(v, 10, 64)
					if err != nil {
						return err
					}
					value.SetInt(i)

				case reflect.Bool:
					b, err := strconv.ParseBool(v)
					if err != nil {
						return err
					}
					value.SetBool(b)
				}
			}
		}
	}
	return nil
}

func trimQuotes(s string) string {
	if len(s) > 0 && s[0] == '"' {
		s = s[1:]
	}
	if len(s) > 0 && s[len(s)-1] == '"' {
		s = s[:len(s)-1]
	}
	return s
}

func (s *EnvSource) Watch(done chan bool, group *sync.WaitGroup) {
	<-done
}

func (s *EnvSource) SetTargetStruct(i interface{}) {
	s.TargetStruct = i
}

func (s *EnvSource) GetTagIds() map[string]string {
	return s.TagIds
}
