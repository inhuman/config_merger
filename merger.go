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
}

func NewMerger() *Merger {
	return &Merger{}
}

func (m *Merger) AddTargetConfigStruct(s interface{}) {
	m.TargetConfigStruct = s
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

//TODO: add kv watcher
//TODO: add json watcher
