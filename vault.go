package config_merger

import (
	"github.com/hashicorp/vault/api"
	"net/http"
	"encoding/json"
)

type VaultSource struct {
	Address      string
	Token        string
	Prefix       string
	HttpClient   *http.Client
	TargetStruct interface{}
}

func (j *VaultSource) Load() error {

	config := api.DefaultConfig()

	if j.HttpClient != nil {
		config.HttpClient = j.HttpClient
	}

	client, err := api.NewClient(config)

	if err != nil {
		return err
	}

	tokenHeader := http.Header{}
	tokenHeader.Set("X-Vault-Token", j.Token)

	client.SetAddress(j.Address)
	client.SetHeaders(tokenHeader)

	secret, err := client.Logical().Read(j.Prefix)

	if err != nil {
		return err
	}

	if secret != nil {

		b, err := json.Marshal(secret.Data)
		if err != nil {
			return err
		}

		err = json.Unmarshal(b, j.TargetStruct)
		if err != nil {
			return err
		}
	}

	return nil
}

func (j *VaultSource) SetTargetStruct(s interface{}) {
	j.TargetStruct = s
}

func (j *VaultSource) SetHttpClient(httpClient *http.Client) {
	j.HttpClient = httpClient
}

func (j *VaultSource) Watch() {}