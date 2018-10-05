package config_merger

import (
	"github.com/fatih/structs"
	"github.com/hashicorp/go-multierror"
	"reflect"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
	"net"
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
		go s.Watch(doneMap[i], &wg)
	}

	if errAll != nil {
		if len(errAll.Errors) > 0 {
			return errAll
		}
	}
	<- m.done

	for d := range m.Sources {
		doneMap[d] <- true
	}
	wg.Wait()
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

	err := m.checkRequiredFields()

	if err != nil {
		return err
	}

	return nil
}

func (m *Merger) GetFinalConfig() map[string]interface{} {
	return structs.Map(m.TargetConfigStruct)
}

func (m *Merger) PrintConfig() {

	t := reflect.TypeOf(m.TargetConfigStruct).Elem()
	v := reflect.ValueOf(m.TargetConfigStruct).Elem()

	fmt.Println(reflect.TypeOf(m.TargetConfigStruct))

	processPrint(t, v, "  ")

}

func processPrint(t reflect.Type, v reflect.Value, offset string) {

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		if field.Type.Kind() == reflect.Struct {
			fmt.Println(offset + field.Name)
			processPrint(field.Type, value, offset + "  ")

			} else {

			column := field.Tag.Get("show_last_symbols")
			if column != "" {

				maskLast, err := strconv.Atoi(column)
				if err == nil {
					fmt.Println(offset + field.Name + ": " + maskString(value.String(), maskLast))
				} else {
					fmt.Println(err)
				}

			} else {
				fmt.Println(offset + field.Name + ": " + value.String())
			}
		}
	}
}

func maskString(s string, showLastSymbols int) string {
	if len(s) <= showLastSymbols {
		return s
	}
	return strings.Repeat("*", len(s)-showLastSymbols) + s[len(s)-showLastSymbols:]
}


func (m *Merger) StopDisconnectTimeout(address string, timeout time.Duration) {

	go func() {
		for {
			conn, err := net.Dial("tcp", address)
			if err != nil {
				fmt.Errorf("TCP error : %s", err.Error())
			}
			if conn == nil {
				fmt.Println("can not reach server")
				m.StopWatch()
			}
			<- time.After(timeout * time.Second)
		}
	}()
}