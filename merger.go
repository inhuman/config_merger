package configMerger

import (
	"github.com/fatih/structs"
	"github.com/hashicorp/go-multierror"
)

type Merger struct {
	Sources            []Source
	TargetConfigStruct interface{}
}

type Source interface {
	Load() error
	SetTargetStruct(s interface{})
	Watch(done chan bool)
}

func NewMerger(s interface{}) *Merger {
	m := &Merger{}
	m.TargetConfigStruct = s
	return m
}

func (m *Merger) AddSource(src Source) {
	src.SetTargetStruct(m.TargetConfigStruct)
	m.Sources = append(m.Sources, src)
}

func (m *Merger) MergeConfigs() error {

	var errAll *multierror.Error

	for _, s := range m.Sources {
		err := s.Load()
		if err != nil {
			errAll = multierror.Append(errAll, err)
		}
		done := make(chan bool)

		go s.Watch(done)
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
