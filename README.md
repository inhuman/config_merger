# configMerger
Config tool for merge configs from different sources. Nested configs not supports

Now supports:
- json file
- consul kv
- vault



### Usage

```golang
package configMerger

import "fmt"

type Cnf struct {
	Message string `json:"message"`
}


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