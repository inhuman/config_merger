package config_merger

import (
	"testing"
	"os"
	"github.com/stretchr/testify/assert"
)

func TestJsonSource_LoadSuccess(t *testing.T) {

	fh := createFileForTest(t, `{"message": "from json"}`)
	path := fh.Name()
	defer func() {
		fh.Close()
		os.Remove(path)
	}()

	jsonSource := &JsonSource{Path: path}

	cnf := &Cnf{}
	jsonSource.SetTargetStruct(cnf)

	err := jsonSource.Load()
	assert.NoError(t, err)

	assert.Equal(t, &Cnf{Message: "from json"}, cnf)

}

func TestJsonSource_LoadWrongJson(t *testing.T) {

	fh := createFileForTest(t, `{"message": "from json",}`)
	path := fh.Name()
	defer func() {
		fh.Close()
		os.Remove(path)
	}()

	jsonSource := &JsonSource{Path: path}
	cnf := &Cnf{}
	jsonSource.SetTargetStruct(cnf)
	err := jsonSource.Load()
	assert.Error(t, err)
}