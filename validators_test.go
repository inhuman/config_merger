package config_merger

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

type TestStruct struct {
	RequiredField string `json:"required_field" required:"true"`
	Empty         string `json:"empty"`
}


//TODO: fix test

func TestCheckRequiredFieldsSuccess(t *testing.T) {


	fh := createFileForTest(t, `{"required_field": "from json"}`)
	path := fh.Name()
	defer func() {
		fh.Close()
		os.Remove(path)
	}()

	jsonSource := &JsonSource{Path: path}
	cnf := &TestStruct{}

	m := NewMerger(cnf)
	m.AddSource(jsonSource)
	err := m.Run()

	assert.NoError(t, err)
	assert.Equal(t, &TestStruct{RequiredField: "from json"}, cnf)

}

func TestCheckRequiredFieldsFail(t *testing.T) {

	fh := createFileForTest(t, `{"message": "from json"}`)
	path := fh.Name()
	defer func() {
		fh.Close()
		os.Remove(path)
	}()

	jsonSource := &JsonSource{Path: path}

	cnf := &TestStruct{}

	m := NewMerger(cnf)
	m.AddSource(jsonSource)

	err := m.Run()

	assert.Error(t, err)

}
