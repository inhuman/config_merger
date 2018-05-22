package configMerger

import (
	"io/ioutil"
	"encoding/json"
	"fmt"
)

type JsonSource struct {
	Path string
}

func (j *JsonSource) Load(s interface{}) {

	file, err := ioutil.ReadFile(j.Path)

	if err != nil {
		fmt.Println(err)
	}

	err = json.Unmarshal([]byte(file), s)
	if err != nil {
		fmt.Println(err)
	}
}

