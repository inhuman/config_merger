package configMerger

import (
	"testing"
	"os"
	"github.com/stretchr/testify/assert"
)

func TestJsonSource_Load(t *testing.T) {

	fh := createFileForTest(t, `{"message": "from json"}`)
	path := fh.Name()
	defer func() {
		fh.Close()
		os.Remove(path)
	}()

	jsonSource := &JsonSource{Path: path}

	cnf := &Cnf{}

	jsonSource.Load(cnf)

	assert.Equal(t, &Cnf{Message: "from json"}, cnf)

}
