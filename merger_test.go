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
	merger := NewMerger()

	fh := createFileForTest(t)
	path := fh.Name()
	defer func() {
		fh.Close()
		os.Remove(path)
	}()

	// add json source
	jsonSource := &JsonSource{
		Path: path,
	}
	merger.AddSource(jsonSource)

	assert.Equal(t, 1, len(merger.Sources))
}

func createFileForTest(t *testing.T) *os.File {
	data := []byte(`{"message": "from json"}`)
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
