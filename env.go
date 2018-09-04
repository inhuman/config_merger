package config_merger

import (
	"os"
	"reflect"
	"sync"
)

type EnvSource struct {
	Variables    []string
	TargetStruct interface{}
	WatchHandler func()
}

func (e *EnvSource) Load() error {

	t := reflect.TypeOf(e.TargetStruct).Elem()
	v := reflect.ValueOf(e.TargetStruct).Elem()

	processEnvTags(t, v)

	return nil
}

func processEnvTags(t reflect.Type, v reflect.Value) error {

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		if field.Type.Kind() == reflect.Struct {
			err := processEnvTags(field.Type, value)
			if err != nil {
				return err
			}
			continue
		}

		column := field.Tag.Get("env")

		if column != "" {
			os.Getenv(column)

			//TODO: add int and float types, just in case
			value.SetString(os.Getenv(column))
		}
	}
	return nil
}


func (e *EnvSource) Watch(done chan bool, group *sync.WaitGroup) {
	<-done
}

func (e *EnvSource) SetTargetStruct(i interface{}) {
	e.TargetStruct = i
}
