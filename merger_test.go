package config_merger

import (
	"testing"
	"os"
	"path/filepath"
	"fmt"
	"time"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
	"net/http"
)

type Cnf struct {
	Message string `json:"message"`
}

type NestedConf struct {
	NestedPassword struct {
		Pass1 string `json:"pass_1" vault:"pass_1"`
		Pass2 string `json:"pass_2" vault:"pass_2"`
	} `json:"nested_password"`
}

func TestMerger_AddSource(t *testing.T) {
	merger := NewMerger(&Cnf{})

	fh := createFileForTest(t, `{"message": "from json"}`)
	path := fh.Name()
	defer func() {
		fh.Close()
		os.Remove(path)
	}()

	jsonSource := &JsonSource{
		Path: path,
	}
	merger.AddSource(jsonSource)

	assert.Equal(t, 1, len(merger.Sources))
}

func TestMerger_AddTargetConfigStruct(t *testing.T) {
	merger := NewMerger(&Cnf{})
	assert.Equal(t, &Cnf{}, merger.TargetConfigStruct)
}

func TestMerger_GetFinalConfig(t *testing.T) {
	merger := NewMerger(&Cnf{})
	assert.Equal(t, map[string]interface{}{"Message": ""}, merger.GetFinalConfig())
}

func TestMerger_Run(t *testing.T) {
	cnf := &Cnf{}
	merger := NewMerger(cnf)

	fh1 := createFileForTest(t, `{"message": "from json"}`)
	path1 := fh1.Name()
	defer func() {
		fh1.Close()
		os.Remove(path1)
	}()
	merger.AddSource(&JsonSource{Path: path1})

	fh2 := createFileForTest(t, `{"message": "from json 2"}`)
	path2 := fh2.Name()
	defer func() {
		fh1.Close()
		os.Remove(path2)
	}()
	merger.AddSource(&JsonSource{Path: path2})

	merger.Run()

	assert.Equal(t, map[string]interface{}{"Message": "from json 2"}, merger.GetFinalConfig())
}

func TestMerger_WrongType(t *testing.T) {
	cnf := Cnf{}
	var panicMessage = "must provide pointer to struct, received [config_merger.Cnf]"

	defer func() {
		if r := recover(); r != nil {
			if r != panicMessage {
				t.Error("Panic message expected '" + panicMessage + "', got '" + r.(string) + "'")
			}
		}
	}()
	NewMerger(cnf)
	t.Error("The code did not panic")
}

func createFileForTest(t *testing.T, s string) *os.File {
	data := []byte(s)
	path := filepath.Join(os.TempDir(), fmt.Sprintf("file.%d", time.Now().UnixNano()))
	fh, err := os.Create(path)
	if err != nil {
		t.Error(err)
	}
	_, err = fh.Write(data)
	if err != nil {
		t.Error(err)
	}

	return fh
}

type Config struct {
	Redis      Redis      `json:"redis"`
	SymbolsMap SymbolsMap `json:"symbols"`
}

type Redis struct {
	Host          string `json:"host"`
	Port          string `json:"port"`
	ChannelPrefix string `json:"channel_prefix"`
	Password      string `json:"password" vault:"redis_password"`
}

type SymbolStats struct {
	RedisChannel string `json:"redis_channel"`
	Otc          int    `json:"otc"`
	Digits       int    `json:"digits"`
}

type SymbolsMap map[string]SymbolStats

func TestMerger_FullTest(t *testing.T) {

	conf := &Config{}

	m := NewMerger(conf)

	// mock for json file
	fh := createFileForTest(t, `{
  "redis": {
    "host": "redis.example.local",
    "port": "5001",
    "channel_prefix": "prod_"
  }
}`)
	path := fh.Name()
	defer func() {
		fh.Close()
		os.Remove(path)
	}()

	// mock for vault
	gock.New("http://vault.example.local").
		Get("/v1/secret/test/config_merger").
		Reply(200).
		BodyString(`{
    "request_id": "fde655e4-d755-6cad-10e5-c0ed790b66ed",
    "lease_id": "",
    "renewable": false,
    "lease_duration": 2764800,
    "data": {
		"redis_password": "redis_strong_password"
	},
    "wrap_info": null,
    "warnings": null,
    "auth": null
}`)

	gock.New("http://consul.example.local:8500").
		Get("/test/prefix").
		Reply(200).
		BodyString(`[
    {
        "LockIndex": 0,
        "Key": "conf_merger/json",
        "Flags": 0,
        "Value": "ewoic3ltYm9scyI6IHsKICAgICJVU0RSVUIiIDogeyJkaWdpdHMiOjUsICJyZWRpc19jaGFubmVsIjoicmVkdXM6Ly9VU0RSVUIiLCAib3RjIjoxfSwKICAgICJVU0ROT0siIDogeyJkaWdpdHMiOjQsICJyZWRpc19jaGFubmVsIjoicmVkaXM6Ly9VU0ROT0siLCAib3RjIjowfSwKICAgICJVU0RKUFkiIDogeyJkaWdpdHMiOjMsICJyZWRpc19jaGFubmVsIjoicmVkaXM6Ly9VU0RKUFkiLCAib3RjIjoxfQp9fQ==",
        "CreateIndex": 463724,
        "ModifyIndex": 464945
    }
]`)

	defer gock.Off()

	kvSource := &ConsulKvJsonSource{
		Address:    "consul.example.local:8500",
		Datacenter: "test",
		Prefix:     "test/prefix",
	}

	client := &http.Client{Transport: &http.Transport{}, Timeout: time.Microsecond}
	gock.InterceptClient(client)

	kvSource.SetHttpClient(client)

	vaultSource := &VaultSource{
		Address: "http://vault.example.local",
		Token:   "dummy_token",
		Prefix:  "secret/test/config_merger",
	}

	client2 := &http.Client{Transport: &http.Transport{}, Timeout: time.Microsecond}
	gock.InterceptClient(client2)
	vaultSource.SetHttpClient(client2)

	m.AddSource(&JsonSource{
		Path: path,
	})
	m.AddSource(vaultSource)
	m.AddSource(kvSource)

	m.Run()

	assert.Equal(t, "redis.example.local", conf.Redis.Host)
	assert.Equal(t, "5001", conf.Redis.Port)
	assert.Equal(t, "prod_", conf.Redis.ChannelPrefix)
	assert.Equal(t, "redis_strong_password", conf.Redis.Password)
	assert.Equal(t, "redus://USDRUB", conf.SymbolsMap["USDRUB"].RedisChannel)
	assert.Equal(t, 0, conf.SymbolsMap["USDNOK"].Otc)
}
