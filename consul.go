package configMerger

import (
	"github.com/hashicorp/consul/api"
	"encoding/json"
	"net/http"
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
