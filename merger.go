package configMerger

type Merger struct {
	Sources []Source
	TargetConfigStruct interface{}
}


type Source interface {
	Load(s interface{})
}

func NewMerger() *Merger {
	return &Merger{}

}

func (m *Merger) AddTargetConfigStruct(s interface{}) {
	m.TargetConfigStruct = s
}

func (m *Merger) AddSource(src Source) {
	m.Sources = append(m.Sources, src)
}

func (m *Merger) MergeConfigs() {
	for _, s := range m.Sources {
		s.Load(m.TargetConfigStruct)
	}
}


func (m *Merger) GetFinalConfig() interface{} {
	return m.TargetConfigStruct
}