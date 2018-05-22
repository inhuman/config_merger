package configMerger

import "fmt"

type TargetConfig struct {
	Message string `json:"message"`
}


func main()  {

	// create config struct
	strct := &TargetConfig{}

	merger := NewMerger()

	merger.AddTargetConfigStruct(strct)

	// add json source
	jsonSource := &JsonSource{
		Path: "/path/to/json",
	}
	merger.AddSource(jsonSource)

	// add consul kv source
	consulKvSource := &KvSource{
		Address: "consul.addr.local:8500",
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

	fmt.Println(merger.GetFinalConfig())
}