package config_merger

import (
	"github.com/hashicorp/vault/api"
	"net/http"
	"encoding/json"
	"reflect"
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

		fillFieldsByTags(j.TargetStruct, b)
	}

	return nil
}

func processTags(t reflect.Type, v reflect.Value, b []byte) error {

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		if field.Type.Kind() == reflect.Struct {
			err := processTags(field.Type, value, b)
			if err != nil {
				return err
			}
			continue
		}

		column := field.Tag.Get("vault")

		if column != "" {
			var bm map[string]string
			bm = make(map[string]string)
			err := json.Unmarshal(b, &bm)
			if err != nil {
				return nil
			}

			//TODO: add int and float types, just in case
			value.SetString(bm[column])
		}
	}

	return nil
}

// receive pointer to struct
func fillFieldsByTags(i interface{}, b []byte) {

	t := reflect.TypeOf(i).Elem()
	v := reflect.ValueOf(i).Elem()

	processTags(t, v, b)
}

func (j *VaultSource) SetTargetStruct(s interface{}) {
	j.TargetStruct = s
}

func (j *VaultSource) SetHttpClient(httpClient *http.Client) {
	j.HttpClient = httpClient
}

func (j *VaultSource) Watch(done chan bool, group *CountWg) {}
