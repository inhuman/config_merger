package configMerger

import (
	"fmt"
	"github.com/hashicorp/consul/api"
	"encoding/json"
)

type KvSource struct {
	Address string //
	Datacenter string
	Prefix string // like this "prefix" (without the /)
}


func (j *KvSource) Load(s interface{}) {

	cnf := api.DefaultConfig()
	cnf.Address = j.Address
	cnf.Datacenter = j.Datacenter

	Client, err := api.NewClient(cnf)

	if err != nil {
		fmt.Println(err)
	}

	kvPair, _, err := Client.KV().Get(j.Prefix, nil)

	if err != nil {
		fmt.Println(err)
	}

	err = json.Unmarshal([]byte(kvPair.Value), s)
	if err != nil {
		fmt.Println(err)
	}
}
