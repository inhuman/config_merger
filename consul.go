package configMerger

import (
	"github.com/hashicorp/consul/api"
	"encoding/json"
	"net/http"
	"github.com/hashicorp/consul/watch"
	"time"
	"fmt"
)

type KvSource struct {
	Address      string
	Datacenter   string
	Prefix       string // like this "prefix" (without the /)
	HttpClient   *http.Client
	WatchHandler func(i interface{})
	TargetStruct interface{}
}

func (j *KvSource) Load() error {

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

	err = json.Unmarshal([]byte(kvPair.Value), j.TargetStruct)
	if err != nil {
		return err
	}

	return nil
}

func (j *KvSource) SetTargetStruct(s interface{}) {
	j.TargetStruct = s
}

func (j *KvSource) SetHttpClient(httpClient *http.Client) {
	j.HttpClient = httpClient
}

func (j *KvSource) Watch() error {

	wp, err := watch.Parse(map[string]interface{}{"type": "keyprefix", "prefix": j.Prefix})
	if err != nil {
		return err
	}

	wp.Datacenter = j.Datacenter
	wp.Handler = j.handle

	for {
		wp.Run(j.Address)
		time.Sleep(time.Second)
	}

	//resultCh := make(chan api.KVPairs, 1)

	//go func() {
	//	for {
	//		select {
	//			case m := <-resultCh:
	//
	//				fmt.Println()
	//
	//				for k, v := range m {
	//					fmt.Printf("key: %v, value: %s", k, v.Value)
	//				}
	//		}
	//	}
	//}()

	return nil
}

func (j *KvSource) handle(u uint64, i interface{}) {

	if i == nil {
		return
	}

	_, ok := i.(api.KVPairs)
	if !ok {
		return
	}

	j.Load()
	fmt.Printf("%+v\n", i)

}

//func (j *KvSource) wrapper(h func(u uint64, i interface{})) func(u uint64, i interface{}) {
//
//	fmt.Println("Wrapper reload the config")
//
//	j.Load()
//	j.WatchHandler(j.TargetStruct)
//
//	return h
//}
