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
	Address      string
	Datacenter   string
	Prefix       string // like this "prefix" (without the /)
	HttpClient   *http.Client
	WatchHandler func()
	TargetStruct interface{}
	Timeout      time.Duration // timeout if disconnect exit
}

func (ckv *ConsulKvSource) Load() error {

	cnf := api.DefaultConfig()
	cnf.Address = ckv.Address
	cnf.Datacenter = ckv.Datacenter

	if ckv.HttpClient != nil {
		cnf.HttpClient = ckv.HttpClient
	}

	client, err := api.NewClient(cnf)

	if err != nil {
		return err
	}

	configMap, err := consul_kv_mapper.BuildMap(client, ckv.Prefix)

	if err != nil {
		return err
	}

	ckv.buildConfig(configMap)

	return nil
}

func (ckv *ConsulKvSource) buildConfig(configMap *consul_kv_mapper.MapType) {

	t := reflect.TypeOf(ckv.TargetStruct).Elem()
	v := reflect.ValueOf(ckv.TargetStruct).Elem()

	processConsulTags(t, v, configMap)

}

func processConsulTags(t reflect.Type, v reflect.Value, configMap *consul_kv_mapper.MapType) error {

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		if field.Type.Kind() == reflect.Struct {
			err := processConsulTags(field.Type, value, configMap)
			if err != nil {
				return err
			}
			continue
		}

		column := field.Tag.Get("consul")

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

	if configMap.Value != "" {
		v.SetString(string(configMap.Value))
	} else {
		processPath(configMap.Children[consul_kv_mapper.KeyType(columnSlice[0])], columnSlice[1:], v)
	}

}

func (ckv *ConsulKvSource) SetTargetStruct(s interface{}) {
	ckv.TargetStruct = s
}

func (ckv *ConsulKvSource) SetHttpClient(httpClient *http.Client) {
	ckv.HttpClient = httpClient
}

func (ckv *ConsulKvSource) Watch(done chan bool, group *sync.WaitGroup) {

	if ckv.WatchHandler != nil {
		wp, err := watch.Parse(map[string]interface{}{"type": "keyprefix", "prefix": ckv.Prefix})

		if err != nil {
			return
		}

		wp.Datacenter = ckv.Datacenter
		wp.Handler = func(u uint64, i interface{}) {
			group.Add(1)
			ckv.handle(u, i)
			group.Done()
		}

		go func() {
			err = wp.Run(ckv.Address)
			if err != nil {
				fmt.Println(err)
			}
		}()
		<-done
		wp.Stop()
	}
}

func (ckv *ConsulKvSource) handle(u uint64, i interface{}) {

	if i == nil {
		return
	}

	_, ok := i.(api.KVPairs)
	if !ok {
		return
	}

	err := ckv.Load()
	if err == nil {
		ckv.WatchHandler()
	} else {
		fmt.Println(err)
	}
}
