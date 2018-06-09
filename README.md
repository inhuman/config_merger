# Config merger
[![Build Status](https://travis-ci.org/inhuman/config_merger.svg?branch=master)](https://travis-ci.org/inhuman/config_merger)

Config tool for merge configs from different sources. 
Now supports:
- json file (simple)
- consul kv (simple, watch)
- vault (simple)


### Usage
##### Simple Mode
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
	NestedPassword struct{
		Pass1 string `json:"pass_1" vault:"pass_1"`
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


	// add consul kv source
	consulKvSource := &configMerger.KvSource{
		Address: "consul.addr:8500",
		Prefix: "conf_merger/json",
		Datacenter: "experiments",
	}
	merger.AddSource(consulKvSource)


	// add vault source
	vaultSource := &configMerger.VaultSource{
		Address: "http://vault.addr:8200",
		Prefix: "secret/test/config_merger",
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

	// add json source
	jsonSource := &configMerger.JsonSource{
		Path: "/path/to/json",
		WatchHandler: func() {
            fmt.Println("json watcher")
            watcher(confStruct)
        },
	}
	merger.AddSource(jsonSource)

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

	// add vault source
	vaultSource := &configMerger.VaultSource{
		Address: "http://vault.addr:8200",
		Prefix: "secret/test/config_merger",
		Token: "vault_token",
	}
	merger.AddSource(vaultSource)

	merger.RunWatch()
}

func watcher(confStruct *Config) {

	fmt.Printf("Message: %s\n", confStruct.Message)
	fmt.Printf("IntValue: %v\n", confStruct.IntValue)
	fmt.Printf("FloatValue: %v\n", confStruct.FloatValue)
	fmt.Printf("Login: %s\n", confStruct.Login)
	fmt.Printf("Password: %s\n", confStruct.Password)
	fmt.Println()
}

```
