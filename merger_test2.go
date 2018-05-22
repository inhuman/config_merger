package configMerger

import (
	"testing"
	"os"
	"fmt"
	"path/filepath"
	"time"
)

type Cnf2 struct {
	Message string `json:"message"`
}


func TestMergerFull(t *testing.T) {

	fh := createFileForTest2(t)
	path := fh.Name()
	defer func() {
		fh.Close()
		os.Remove(path)
	}()

	// create config struct
	strct := &Cnf2{}

	merger := NewMerger()

	merger.AddTargetConfigStruct(strct)

	// add json source
	jsonSource := &JsonSource{
		Path: path,
	}
	merger.AddSource(jsonSource)

	// add consul kv source
	consulKvSource := &KvSource{
		Address: "consul.infra1.local:8500",
		Prefix: "conf_merger/json",
		Datacenter: "experiments",
	}
	merger.AddSource(consulKvSource)


	// add vault source
	vaultSource := &VaultSource{
		Address: "http://vault.infra1.local:8200",
		Prefix: "secret/test/config_merger",
		Token: "415084eb-7dda-f0cd-9657-515a93d15962",
	}
	merger.AddSource(vaultSource)

	merger.MergeConfigs()

	fmt.Println(merger.GetFinalConfig())


}

func createFileForTest2(t *testing.T) *os.File {
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