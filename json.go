package configMerger

import (
	"io/ioutil"
	"encoding/json"
)

type JsonSource struct {
	Path string
}

func (j *JsonSource) Load(s interface{}) error {

	file, err := ioutil.ReadFile(j.Path)

	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(file), s)
	if err != nil {
		return err
	}

	return nil
}
