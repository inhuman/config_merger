package config_merger

import (
	"github.com/fatih/structs"
	"github.com/hashicorp/go-multierror"
	"reflect"
	"fmt"
)

type Merger struct {
	Sources            []Source
	TargetConfigStruct interface{}
	Done chan bool
}

type Source interface {
	Load() error
	SetTargetStruct(s interface{})
	Watch()
}

func NewMerger(s interface{}) *Merger {
	m := &Merger{
		Done: make(chan bool),
	}

	if reflect.ValueOf(s).Kind() != reflect.Ptr {
		panic(fmt.Sprintf("must provide pointer to struct, received [%T]", s))
	}

	err := validateStruct(s)
	if err != nil {
		panic(err.Error())
	}

	m.TargetConfigStruct = s
	return m
}

func (m *Merger) AddSource(src Source) {
	src.SetTargetStruct(m.TargetConfigStruct)
	m.Sources = append(m.Sources, src)
}

func (m *Merger) RunWatch() error {

	var errAll *multierror.Error

	for _, s := range m.Sources {
		err := s.Load()
		if err != nil {
			errAll = multierror.Append(errAll, err)
		}
		go s.Watch()
	}

	if errAll != nil {
		if len(errAll.Errors) > 0 {
			return errAll
		}
	}
	<-m.Done

	return nil
}

func (m *Merger) StopWatch() {
	m.Done <- true
}

func (m *Merger) Run() error {

	var errAll *multierror.Error

	for _, s := range m.Sources {
		err := s.Load()
		if err != nil {
			errAll = multierror.Append(errAll, err)
		}
	}

	if errAll != nil {
		if len(errAll.Errors) > 0 {
			return errAll
		}
	}

	return nil
}

func (m *Merger) GetFinalConfig() map[string]interface{} {
	return structs.Map(m.TargetConfigStruct)
}
