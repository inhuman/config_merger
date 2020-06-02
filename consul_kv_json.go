package config_merger

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
	"net/http"
	"sync"
	"time"
)

type ConsulKvJsonSource struct {
	SourceModel
	Address    string
	Datacenter string
	Prefix     string // like this "prefix" (without the /)
	HttpClient *http.Client
	Timeout    time.Duration // timeout if disconnect exit
}

//TODO: implement tag ids in consul kv json

func (s *ConsulKvJsonSource) Load() error {

	cnf := api.DefaultConfig()
	cnf.Address = s.Address
	cnf.Datacenter = s.Datacenter

	if s.HttpClient != nil {
		cnf.HttpClient = s.HttpClient
	}

	Client, err := api.NewClient(cnf)

	if err != nil {
		return err
	}

	kvPair, _, err := Client.KV().Get(s.Prefix, nil)

	if err != nil {
		return err
	}

	if kvPair != nil {
		err = json.Unmarshal([]byte(kvPair.Value), s.TargetStruct)
		if err != nil {
			return err
		}
	} else {
		return errors.New("kv pair is nil")
	}

	return nil
}

func (s *ConsulKvJsonSource) SetTargetStruct(i interface{}) {
	s.TargetStruct = i
}

func (s *ConsulKvJsonSource) SetHttpClient(httpClient *http.Client) {
	s.HttpClient = httpClient
}

func (s *ConsulKvJsonSource) Watch(done chan bool, group *sync.WaitGroup) {

	if s.WatchHandler != nil {
		wp, err := watch.Parse(map[string]interface{}{"type": "keyprefix", "prefix": s.Prefix})

		if err != nil {
			return
		}

		wp.Datacenter = s.Datacenter
		wp.Handler = func(u uint64, i interface{}) {
			group.Add(1)
			s.handle(u, i)
			group.Done()
		}

		go func() {
			err = wp.Run(s.Address)
			if err != nil {
				fmt.Println(err)
			}
		}()
		<-done
		wp.Stop()
	}
}

func (s *ConsulKvJsonSource) handle(u uint64, i interface{}) {

	if i == nil {
		return
	}

	_, ok := i.(api.KVPairs)
	if !ok {
		return
	}

	err := s.Load()
	if err == nil {
		s.WatchHandler()
	} else {
		fmt.Println(err)
	}
}

func (s *ConsulKvJsonSource) GetTagIds() map[string]string {
	return s.TagIds
}
