package config_merger

import (
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
	"net/http"
	"testing"
	"time"
)

func TestKvJSONSource_LoadSuccess(t *testing.T) {

	defer gock.Off()

	gock.New("http://consul.example.local:8500").
		Get("/test/prefix").
		Reply(200).
		BodyString(`[
    {
        "LockIndex": 0,
        "Key": "conf_merger/json",
        "Flags": 0,
        "Value": "ewogICJNZXNzYWdlIjogImZyb20ga3YiCn0K",
        "CreateIndex": 463724,
        "ModifyIndex": 464945
    }
]`)

	kvSource := &ConsulKvJsonSource{
		Address:    "consul.example.local:8500",
		Datacenter: "test",
		Prefix:     "test/prefix",
	}

	client := &http.Client{Transport: &http.Transport{}, Timeout: time.Microsecond}
	gock.InterceptClient(client)

	kvSource.SetHttpClient(client)

	cnf := &Cnf{}
	kvSource.SetTargetStruct(cnf)
	err := kvSource.Load()

	assert.NoError(t, err)
	assert.Equal(t, "from kv", cnf.Message)
}

func TestKvJSONSource_LoadWrongAddr(t *testing.T) {

	kvSource := &ConsulKvJsonSource{
		Address:    "consul.example.local:8500",
		Datacenter: "test",
		Prefix:     "test/prefix",
	}

	client := &http.Client{Transport: &http.Transport{}, Timeout: time.Microsecond}
	gock.InterceptClient(client)

	kvSource.SetHttpClient(client)

	cnf := &Cnf{}

	kvSource.SetTargetStruct(cnf)

	err := kvSource.Load()
	assert.Error(t, err)
}
