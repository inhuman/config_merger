package config_merger

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/watch"
	"net/http"
	"sync"
	"time"
)

type KvSource struct {
	Address      string
	Datacenter   string
	Prefix       string // like this "prefix" (without the /)
	HttpClient   *http.Client
	WatchHandler func()
	TargetStruct interface{}
	Timeout      time.Duration // timeout if disconnect exit
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

	if kvPair != nil {
		err = json.Unmarshal([]byte(kvPair.Value), j.TargetStruct)
		if err != nil {
			return err
		}
	} else {
		return errors.New("kv pair is nil")
	}

	return nil
}

func (j *KvSource) SetTargetStruct(s interface{}) {
	j.TargetStruct = s
}

func (j *KvSource) SetHttpClient(httpClient *http.Client) {
	j.HttpClient = httpClient
}

func (j *KvSource) Watch(done chan bool, group *sync.WaitGroup) {

	if j.WatchHandler != nil {
		wp, err := watch.Parse(map[string]interface{}{"type": "keyprefix", "prefix": j.Prefix})

		if err != nil {
			return
		}

		wp.Datacenter = j.Datacenter
		wp.Handler = func(u uint64, i interface{}) {
			group.Add(1)
			j.handle(u, i)
			group.Done()
		}

		go func() {
			err = wp.Run(j.Address)
			if err != nil {
				fmt.Println(err)
			}
		}()
		<-done
		wp.Stop()
	}
}

func (j *KvSource) handle(u uint64, i interface{}) {

	if i == nil {
		return
	}

	_, ok := i.(api.KVPairs)
	if !ok {
		return
	}

	err := j.Load()
	if err == nil {
		j.WatchHandler()
	} else {
		fmt.Println(err)
	}
}
