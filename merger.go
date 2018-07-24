package config_merger

import (
	"github.com/fatih/structs"
	"github.com/hashicorp/go-multierror"
	"reflect"
	"fmt"
	"sync"
	"time"
	"sync/atomic"
)

type Merger struct {
	Sources            []Source
	TargetConfigStruct interface{}
	done               chan bool
	wg                 CountWg
}

type Source interface {
	Load() error
	SetTargetStruct(s interface{})
	Watch(done chan bool, group *CountWg)
}

type CountWg struct {
	sync.WaitGroup
	Count int32
}

func (cg CountWg) Add(delta int) {
	atomic.AddInt32(&cg.Count, -1)
	cg.WaitGroup.Add(delta)
}

// Decrement on Done
func (cg CountWg) Done() {
	atomic.AddInt32(&cg.Count, -1)
	cg.WaitGroup.Done()
}

func (cg CountWg) DoneAll() {
	for i := 0; int32(i) <= cg.Count; i++ {
		cg.Done()
	}
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


	doneMap := make(map[int]chan bool)

	for i, s := range m.Sources {
		err := s.Load()
		if err != nil {
			errAll = multierror.Append(errAll, err)
		}
		doneMap[i] = make(chan bool)
		go s.Watch(doneMap[i], &m.wg)
	}

	if errAll != nil {
		if len(errAll.Errors) > 0 {
			return errAll
		}
	}

	<-m.done

	for d := range m.Sources {
		doneMap[d] <- true
	}
	m.wg.Wait()

	return nil
}

func (m *Merger) StopWatch(timeout time.Duration) {

	//TODO: fix timeout

	if timeout == 0 {
		m.done <- true
	} else {
		<- time.After(timeout)
		m.wg.DoneAll()
	}
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
