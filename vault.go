package configMerger

import (
	"github.com/hashicorp/vault/api"
	"net/http"
	"encoding/json"
)

type VaultSource struct {
	Address    string
	Token      string
	Prefix     string
	HttpClient *http.Client
}

func (j *VaultSource) Load(s interface{}) error {

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

		err = json.Unmarshal(b, s)
		if err != nil {
			return err
		}
	}

	return nil
}

func (j *VaultSource) SetHttpClient(httpClient *http.Client) {
	j.HttpClient = httpClient
}
