package config_merger

import (
	"fmt"
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/watch"
	"github.com/inhuman/consul-kv-mapper"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"
)

type ConsulKvSource struct {
	SourceModel
	Address    string
	Datacenter string
	Prefix     string // like this "prefix" (without the /)
	HttpClient *http.Client
	Timeout    time.Duration // timeout if disconnect exit
}

func (s *ConsulKvSource) Load() error {

	cnf := api.DefaultConfig()
	cnf.Address = s.Address
	cnf.Datacenter = s.Datacenter

	if s.HttpClient != nil {
		cnf.HttpClient = s.HttpClient
	}

	client, err := api.NewClient(cnf)

	if err != nil {
		return err
	}

	configMap, err := consul_kv_mapper.BuildMap(client, s.Prefix)

	if err != nil {
		return err
	}

	if err := s.buildConfig(configMap); err != nil {
		return err
	}

	return nil
}

func (s *ConsulKvSource) buildConfig(configMap *consul_kv_mapper.MapType) error {

	t := reflect.TypeOf(s.TargetStruct).Elem()
	v := reflect.ValueOf(s.TargetStruct).Elem()

	if err := s.processConsulTags(t, v, configMap); err != nil {
		return err
	}
	return nil
}

func (s *ConsulKvSource) processConsulTags(t reflect.Type, v reflect.Value, configMap *consul_kv_mapper.MapType) error {

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		if field.Type.Kind() == reflect.Struct {
			err := s.processConsulTags(field.Type, value, configMap)
			if err != nil {
				return err
			}
			continue
		}

		column := GetTagContents(s, "consul", field)

		if column != "" {
			columnSlice := strings.Split(column, "/")
			child := configMap.Children[consul_kv_mapper.KeyType(columnSlice[0])]

			if child != nil {
				if len(columnSlice) < 2 {
					//TODO: add int and float types, just in case
					value.SetString(string(child.Value))
				} else {
					processPath(child, columnSlice[1:], value)
				}
			}
		}
	}
	return nil
}

func processPath(configMap *consul_kv_mapper.MapType, columnSlice []string, v reflect.Value) {

	if len(configMap.Children) < 1 {
		v.SetString(string(configMap.Value))
	} else {
		processPath(configMap.Children[consul_kv_mapper.KeyType(columnSlice[0])], columnSlice[1:], v)
	}

}

func (s *ConsulKvSource) SetTargetStruct(i interface{}) {
	s.TargetStruct = i
}

func (s *ConsulKvSource) SetHttpClient(httpClient *http.Client) {
	s.HttpClient = httpClient
}

func (s *ConsulKvSource) Watch(done chan bool, group *sync.WaitGroup) {

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

func (s *ConsulKvSource) handle(u uint64, i interface{}) {

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

func (s *ConsulKvSource) GetTagIds() map[string]string {
	return s.TagIds
}
