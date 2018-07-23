package config_merger

import (
	"github.com/fatih/structs"
	"github.com/hashicorp/go-multierror"
	"reflect"
	"fmt"
	"sync"
)

type Merger struct {
	Sources            []Source
	TargetConfigStruct interface{}
	done chan bool
}

type Source interface {
	Load() error
	SetTargetStruct(s interface{})
	Watch(done chan bool, group *sync.WaitGroup)
}

func NewMerger(s interface{}) *Merger {
	m := &Merger{
		done: make(chan bool),
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

	var wg sync.WaitGroup

	doneMap := make(map[int]chan bool)

	for i, s := range m.Sources {
		err := s.Load()
		if err != nil {
			errAll = multierror.Append(errAll, err)
		}

		doneMap[i] = make(chan bool)
		fmt.Println("Done chan created for watcher", i)

		go s.Watch(doneMap[i], &wg)
	}

	if errAll != nil {
		if len(errAll.Errors) > 0 {
			return errAll
		}
	}

	<- m.done
	fmt.Println("Done reseived")

	for d := range m.Sources {
		fmt.Println("Sending done to watcher", d)
		doneMap[d] <- true
		fmt.Println("Sent done to watcher", d)
	}

	fmt.Println("Done sent to watchers, waiting wg")

	wg.Wait()
	fmt.Println("wg wait complete")

	return nil
}

func (m *Merger) StopWatch() {
	m.done <- true
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
