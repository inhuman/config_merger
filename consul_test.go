package configMerger

import (
	"testing"
	"gopkg.in/h2non/gock.v1"
	"net/http"
	"github.com/stretchr/testify/assert"
)

func TestKvSource_LoadSuccess(t *testing.T) {

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

	kvSource := &KvSource{
		Address:    "consul.example.local:8500",
		Datacenter: "test",
		Prefix:     "test/prefix",
	}

	client := &http.Client{Transport: &http.Transport{}}
	gock.InterceptClient(client)

	kvSource.SetHttpClient(client)

	cnf := &Cnf{}
	err := kvSource.Load(cnf)

	assert.NoError(t, err)
	assert.Equal(t, "from kv", cnf.Message)
}

func TestKvSource_LoadWrongAddr(t *testing.T) {

	kvSource := &KvSource{
		Address:    "consul.example.local:8500",
		Datacenter: "test",
		Prefix:     "test/prefix",
	}

	client := &http.Client{Transport: &http.Transport{}}
	gock.InterceptClient(client)

	kvSource.SetHttpClient(client)

	cnf := &Cnf{}
	err := kvSource.Load(cnf)
	assert.Error(t, err)
}