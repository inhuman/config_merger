package configMerger

import (
	"testing"
	"os"
	"path/filepath"
	"fmt"
	"time"
	"github.com/stretchr/testify/assert"
)

type Cnf struct {
	Message string `json:"message"`
}

func TestMerger_AddSource(t *testing.T) {
	merger := NewMerger(&Cnf{})

	fh := createFileForTest(t, `{"message": "from json"}`)
	path := fh.Name()
	defer func() {
		fh.Close()
		os.Remove(path)
	}()

	jsonSource := &JsonSource{
		Path: path,
	}
	merger.AddSource(jsonSource)

	assert.Equal(t, 1, len(merger.Sources))
}

func TestMerger_AddTargetConfigStruct(t *testing.T) {
	merger := NewMerger(&Cnf{})
	assert.Equal(t, &Cnf{}, merger.TargetConfigStruct)
}

func TestMerger_GetFinalConfig(t *testing.T) {
	merger := NewMerger(&Cnf{})
	assert.Equal(t, map[string]interface{}{"Message": ""}, merger.GetFinalConfig())
}

func TestNewMerger(t *testing.T) {
	merger := NewMerger(&Cnf{})

	assert.Equal(t, &Merger{TargetConfigStruct: &Cnf{}}, merger)
}

func TestMerger_Run(t *testing.T) {
	cnf := &Cnf{}
	merger := NewMerger(cnf)

	fh1 := createFileForTest(t, `{"message": "from json"}`)
	path1 := fh1.Name()
	defer func() {
		fh1.Close()
		os.Remove(path1)
	}()
	merger.AddSource(&JsonSource{Path: path1})

	fh2 := createFileForTest(t, `{"message": "from json 2"}`)
	path2 := fh2.Name()
	defer func() {
		fh1.Close()
		os.Remove(path2)
	}()
	merger.AddSource(&JsonSource{Path: path2})


	merger.Run()

	assert.Equal(t, map[string]interface{}{"Message": "from json 2"}, merger.GetFinalConfig())
}

func createFileForTest(t *testing.T, s string) *os.File {
	data := []byte(s)
	path := filepath.Join(os.TempDir(), fmt.Sprintf("file.%d", time.Now().UnixNano()))
	fh, err := os.Create(path)
	if err != nil {
		t.Error(err)
	}
	_, err = fh.Write(data)
	if err != nil {
		t.Error(err)
	}

	return fh
}
