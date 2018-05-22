package configMerger

import (
	"github.com/hashicorp/vault/api"
	"fmt"
	"net/http"
	"encoding/json"
)

type VaultSource struct {
	Address string
	Token   string
	Prefix  string
	Value   string
}

func (j *VaultSource) Load(s interface{}) {

	config := api.DefaultConfig()
	client, err := api.NewClient(config)

	if err != nil {
		fmt.Println(err)
	}

	tokenHeader := http.Header{}
	tokenHeader.Set("X-Vault-Token", j.Token)

	client.SetAddress(j.Address)
	client.SetHeaders(tokenHeader)

	secret, err := client.Logical().Read(j.Prefix)

	if err != nil {
		fmt.Println(err)
	} else {

		if secret != nil {

			b, err := json.Marshal(secret.Data)
			if err != nil {
				fmt.Println(err)
			}

			err = json.Unmarshal(b, s)
			if err != nil {
				fmt.Println(err)
			}

		}
	}
}
