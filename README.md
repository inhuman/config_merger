# Config merger
[![Build Status](https://travis-ci.org/inhuman/config_merger.svg?branch=master)](https://travis-ci.org/inhuman/config_merger)

Config tool for merge configs from different sources. 
Now supports:
- json file (simple)
- consul kv (simple, watch)
- consul kv json (simple, watch)
- vault (simple)
- env (simple)


### Usage
##### Simple Mode
```golang
package main

import (
    "fmt"
	"github.com/inhuman/configMerger"
)

type Config struct {
	Message        string `json:"message" required:"true"`
	IntValue       int    `json:"int_value"`
	FloatValue     string `consul:"nested_password/another_level/one_more/key3"`
	Login          string `json:"login" env:"ENV_LOGIN" consul:"login"`
	Password       string `json:"password" env:"ENV_PASSWORD" consul:"nested_password/pass_1"`
	NestedPassword struct {
		Pass1 string `json:"pass_1" vault:"db/admin:pass_1" env:"ENV_PASSWORD" consul:"nested_password/another_level/key"`
	} `json:"nested_password"`
}

// Override priority depends order of adding - last overrides previous
func main()  {

	// create config struct
	strct := &Config{}

	merger := configMerger.NewMerger(strct)

	// add json source
	jsonSource := &configMerger.JsonSource{
		Path: "/path/to/json",
	}
	merger.AddSource(jsonSource)


	// add consul kv json source
	consulKvJsonSource := &configMerger.KvJsonSource{
		Address: "consul.addr:8500",
		Prefix: "conf_merger/json",
		Datacenter: "experiments",
	}
	merger.AddSource(consulKvJsonSource)

	// add consul kv source
	consulKvJsonSource := &configMerger.KvSource{
		Address: "consul.addr:8500",
		Prefix: "conf_merger/config",
		Datacenter: "experiments",
	}
	merger.AddSource(consulKvSource)

	// add vault source
	vaultSource := &configMerger.VaultSource{
		Address: "http://vault.addr:8200",
		Prefix: "secret/service/",
		Token: "vault_token",
	}
	merger.AddSource(vaultSource)

	merger.Run()

	// There is two ways to get final config

	fmt.Printf("%s", merger.GetFinalConfig()) // return map[string]interface{}

	fmt.Printf("%+v", strct) // return struct
}
```

##### Watch Mode
```golang
package main

import (
    "fmt"
	"github.com/inhuman/configMerger"
)

type Config struct {
	Message    string  `json:"message"`
	IntValue   int     `json:"int_value"`
	FloatValue float32 `json:"float_value"`
	Login      string  `json:"login"`
	Password   string  `json:"password"`
}

// Override priority depends order of adding - last overrides previous
func main()  {

	// create config struct
	strct := &Config{}

	merger := configMerger.NewMerger(strct)

	// add consul kv source
	consulKvSource := &configMerger.KvSource{
		Address: "consul.addr:8500",
		Prefix: "conf_merger/json",
		Datacenter: "experiments",
		WatchHandler: func() {
            fmt.Println("watcher 2")
            watcher(confStruct)
        },
	}
	merger.AddSource(consulKvSource)

    // add consul kv source 2
    kvSource2 := &configMerger.KvSource{
		Address: "consul.addr:8500",
		Prefix: "conf_merger/another_path",
		Datacenter: "experiments",
		WatchHandler: func() {
			fmt.Println("watcher 2")
			watcher(confStruct)
		},
	}
	m.AddSource(kvSource2)

	merger.RunWatch()
}

func watcher(confStruct *Config) {
	fmt.Printf("Message: %s\n", confStruct.Message)
}

```
#### Dynamic field names 

Supported by consul-kv, vault, env sources

Tag has higher priority

```golang

type CnfTag struct {
	Message string `json:"message" tagId:"cnf_tag_message"`
}

vaultSource := &VaultSource{
    Address: "http://vault.example.local",
    Token:   "dummy_token",
    Prefix:  "secret/test/config_merger",
}
	
vaultSource.TagIds = map[string]string{
    "cnf_tag_message": "message",
}
	
```
