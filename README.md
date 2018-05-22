# configMerger
[![Build Status](https://travis-ci.org/inhuman/configMerger.svg?branch=master)](https://travis-ci.org/inhuman/configMerger)

Config tool for merge configs from different sources. Nested configs not supports

Now supports:
- json file
- consul kv
- vault



### Usage

```golang
package configMerger

import "fmt"

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
	strct := &Cnf{}

	merger := NewMerger()

	merger.AddTargetConfigStruct(strct)

	// add json source
	jsonSource := &JsonSource{
		Path: "/path/to/json",
	}
	merger.AddSource(jsonSource)

	// add consul kv source
	consulKvSource := &KvSource{
		Address: "consul.addr:8500",
		Prefix: "conf_merger/json",
		Datacenter: "experiments",
	}
	merger.AddSource(consulKvSource)


	// add vault source
	vaultSource := &VaultSource{
		Address: "http://vault.addr:8200",
		Prefix: "secret/test/config_merger",
		Token: "vault_token",
	}
	merger.AddSource(vaultSource)

	merger.MergeConfigs()

    // There is two ways to get final config

    fmt.Printf("%s", merger.GetFinalConfig()) // return map[string]interface{}

    fmt.Printf("%s", cnf) // return struct
}
```

