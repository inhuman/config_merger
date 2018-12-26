package config_merger

import (
	"encoding/json"
	"github.com/hashicorp/vault/api"
	"github.com/vmware/vic/pkg/errors"
	"net/http"
	"reflect"
	"strings"
	"sync"
)

type VaultSource struct {
	Address      string
	Token        string
	Prefix       string
	HttpClient   *http.Client
	TargetStruct interface{}
	Client       *api.Client
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

	if err := client.SetAddress(j.Address); err != nil {
		return err
	}

	client.SetHeaders(tokenHeader)
	j.Client = client

	err = j.fillFieldsByTags(j.TargetStruct, []byte{})
	if err != nil {
		return err
	}

	return nil
}

func (j *VaultSource) readByPrefix(prefix string) ([]byte, error) {

	fullPrefix := j.Prefix + prefix
	secret, err := j.Client.Logical().Read(fullPrefix)

	if err != nil {
		return nil, err
	}

	if secret != nil {
		return json.Marshal(secret.Data)
	}

	return nil, errors.New("Vault entry " + fullPrefix + " is empty")
}

func (j *VaultSource) processVaultTags(t reflect.Type, v reflect.Value, b []byte) error {

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		if field.Type.Kind() == reflect.Struct {
			err := j.processVaultTags(field.Type, value, b)
			if err != nil {
				return err
			}
			continue
		}

		column := field.Tag.Get("vault")

		field.Tag = ""

		if column != "" {

			prefix, name, err := j.getPrefixAndName(column)
			if err != nil {
				return err
			}

			b, err = j.readByPrefix(prefix)
			if err != nil {
				return err
			}

			var bm map[string]string
			bm = make(map[string]string)
			err = json.Unmarshal(b, &bm)
			if err != nil {
				return nil
			}

			//TODO: add int and float types, just in case
			value.SetString(bm[name])
		}
	}

	return nil
}

func (j *VaultSource) getPrefixAndName(tag string) (string, string, error) {

	tagParts := strings.Split(tag, ":")

	switch {

	case len(tagParts) > 2 || len(tagParts) < 1:
		return "", "", errors.New("Unconditional tag `" + tag + "`")

	case len(tagParts) == 2:
		return tagParts[0], tagParts[1], nil

	case len(tagParts) == 1 && j.Prefix == "":

		prefixParts := strings.Split(tag, "/")
		if len(prefixParts) > 1 {
			return "", "", errors.New("Unconditional tag `" + tag + "` - missing field name")
		}
		return "", "", errors.New("Unconditional tag `" + tag + "` - missing prefix")

	case len(tagParts) == 1 && j.Prefix != "":
		return "", tagParts[0], nil
	default:
		return "", "", errors.New("Parsing `" + tag + "` prefix and field name unknown error")
	}
}

// receive pointer to struct
func (j *VaultSource) fillFieldsByTags(i interface{}, b []byte) error {

	t := reflect.TypeOf(i).Elem()
	v := reflect.ValueOf(i).Elem()

	err := j.processVaultTags(t, v, b)
	if err != nil {
		return err
	}

	return nil
}

func (j *VaultSource) SetTargetStruct(s interface{}) {
	j.TargetStruct = s
}

func (j *VaultSource) SetHttpClient(httpClient *http.Client) {
	j.HttpClient = httpClient
}

func (j *VaultSource) Watch(done chan bool, group *sync.WaitGroup) {
	<-done
}
