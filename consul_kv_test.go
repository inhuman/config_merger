package config_merger

import (
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
	"net/http"
	"testing"
	"time"
)

type ConfigConsul struct {
	SimpleKey string `consul:"login"`
	Key       string `consul:"nested_config/key_1"`
	SubConfig struct {
		SubKey string `consul:"nested_config/another_level/key_2"`
	}
}

func TestKvSource_LoadSuccess(t *testing.T) {

	defer gock.Off()

	gock.New("http://consul.example.local:8500").
		Get("/v1/kv/main_config").
		Reply(200).
		BodyString(`[
    {
        "LockIndex": 0,
        "Key": "main_config/nested_config/another_level/",
        "Flags": 0,
        "Value": null,
        "CreateIndex": 463726,
        "ModifyIndex": 464947
    },
    {
        "LockIndex": 0,
        "Key": "main_config/nested_config/another_level/key_2",
        "Flags": 0,
        "Value": "c3ViIHZhbHVlIDIgZnJvbSBrdg==",
        "CreateIndex": 463726,
        "ModifyIndex": 464947
    },
    {
        "LockIndex": 0,
        "Key": "main_config/login",
        "Flags": 0,
        "Value": "bG9naW4gZnJvbSBrdg==",
        "CreateIndex": 463724,
        "ModifyIndex": 464945
    },
    {
        "LockIndex": 0,
        "Key": "main_config/nested_config/key_1",
        "Flags": 0,
        "Value": "dmFsdWUgMSBmcm9tIGt2",
        "CreateIndex": 463725,
        "ModifyIndex": 464946
    }
]`)


	kvSource := &ConsulKvSource{
		Address:    "consul.example.local:8500",
		Datacenter: "test",
		Prefix:     "main_config",
	}

	client := &http.Client{Transport: &http.Transport{}, Timeout: time.Microsecond}
	gock.InterceptClient(client)

	kvSource.SetHttpClient(client)

	cnf := &ConfigConsul{}
	kvSource.SetTargetStruct(cnf)
	err := kvSource.Load()

	assert.NoError(t, err)
	assert.Equal(t, "login from kv", cnf.SimpleKey)
	assert.Equal(t, "value 1 from kv", cnf.Key)
	assert.Equal(t, "sub value 2 from kv", cnf.SubConfig.SubKey)
}

func TestKvSource_LoadWrongAddr(t *testing.T) {

	//kvSource := &ConsulKvSource{
	//	Address:    "consul.example.local:8500",
	//	Datacenter: "test",
	//	Prefix:     "test/prefix",
	//}
	//
	//client := &http.Client{Transport: &http.Transport{}, Timeout: time.Microsecond}
	//gock.InterceptClient(client)
	//
	//kvSource.SetHttpClient(client)
	//
	//cnf := &Cnf{}
	//
	//kvSource.SetTargetStruct(cnf)
	//
	//err := kvSource.Load()
	//assert.Error(t, err)
}
