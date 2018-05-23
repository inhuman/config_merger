package configMerger

import (
	"github.com/hashicorp/consul/api"
	"encoding/json"
	"net/http"
	"github.com/hashicorp/consul/watch"
	"fmt"
)

type KvSource struct {
	Address    string //
	Datacenter string
	Prefix     string // like this "prefix" (without the /)
	HttpClient *http.Client
}

func (j *KvSource) Load(s interface{}) error {

	cnf := api.DefaultConfig()
	cnf.Address = j.Address
	cnf.Datacenter = j.Datacenter

	if j.HttpClient != nil {
		cnf.HttpClient = j.HttpClient
	}

	Client, err := api.NewClient(cnf)

	if err != nil {
		return err
	}

	kvPair, _, err := Client.KV().Get(j.Prefix, nil)

	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(kvPair.Value), s)
	if err != nil {
		return err
	}

	return nil
}

func (j *KvSource) SetHttpClient(httpClient *http.Client) {
	j.HttpClient = httpClient
}

func (j *KvSource) Watch() error {

	fmt.Println("Running consul kv watcher")

	wp, err := watch.Parse(map[string]interface{}{"type": "keyprefix", "prefix": j.Prefix})
	if err != nil {
		return err
	}

	wp.Datacenter = j.Datacenter

	wp.Handler = func(u uint64, i interface{}) {

		fmt.Println("Watch plan handler")

		if i == nil {
			return
		}

		kvs, ok := i.(api.KVPairs)
		if !ok {
			return
		}

		fmt.Printf("%v", kvs)
	}

	fmt.Printf("%+v\n", wp)

	go wp.Run(j.Address)

	fmt.Println("Exiting..")


	return nil
}

