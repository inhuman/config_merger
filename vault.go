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
	SourceModel
	Address    string
	Token      string
	Prefix     string
	HttpClient *http.Client
	Client     *api.Client
}

func (s *VaultSource) Load() error {

	config := api.DefaultConfig()

	if s.HttpClient != nil {
		config.HttpClient = s.HttpClient
	}

	client, err := api.NewClient(config)
	if err != nil {
		return err
	}

	tokenHeader := http.Header{}
	tokenHeader.Set("X-Vault-Token", s.Token)

	if err := client.SetAddress(s.Address); err != nil {
		return err
	}

	client.SetHeaders(tokenHeader)
	s.Client = client

	err = s.fillFieldsByTags(s.TargetStruct, []byte{})
	if err != nil {
		return err
	}

	return nil
}

func (s *VaultSource) readByPrefix(prefix string) ([]byte, error) {

	fullPrefix := s.Prefix + prefix
	secret, err := s.Client.Logical().Read(fullPrefix)

	if err != nil {
		return nil, err
	}

	if secret != nil {
		return json.Marshal(secret.Data)
	}

	return nil, errors.New("Vault entry " + fullPrefix + " is empty")
}

func (s *VaultSource) processVaultTags(t reflect.Type, v reflect.Value, b []byte) error {

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		if field.Type.Kind() == reflect.Struct {
			err := s.processVaultTags(field.Type, value, b)
			if err != nil {
				return err
			}
			continue
		}

		column := GetTagContents(s, "vault", field)

		if column != "" {
			field.Tag = ""

			prefix, name, err := s.getPrefixAndName(column)
			if err != nil {
				return err
			}

			b, err = s.readByPrefix(prefix)
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

func (s *VaultSource) getPrefixAndName(tag string) (string, string, error) {

	tagParts := strings.Split(tag, ":")

	switch {

	case len(tagParts) > 2 || len(tagParts) < 1:
		return "", "", errors.New("Unconditional tag `" + tag + "`")

	case len(tagParts) == 2:
		return tagParts[0], tagParts[1], nil

	case len(tagParts) == 1 && s.Prefix == "":

		prefixParts := strings.Split(tag, "/")
		if len(prefixParts) > 1 {
			return "", "", errors.New("Unconditional tag `" + tag + "` - missing field name")
		}
		return "", "", errors.New("Unconditional tag `" + tag + "` - missing prefix")

	case len(tagParts) == 1 && s.Prefix != "":
		return "", tagParts[0], nil
	default:
		return "", "", errors.New("Parsing `" + tag + "` prefix and field name unknown error")
	}
}

// receive pointer to struct
func (s *VaultSource) fillFieldsByTags(i interface{}, b []byte) error {

	t := reflect.TypeOf(i).Elem()
	v := reflect.ValueOf(i).Elem()

	err := s.processVaultTags(t, v, b)
	if err != nil {
		return err
	}

	return nil
}

func (s *VaultSource) SetTargetStruct(i interface{}) {
	s.TargetStruct = i
}

func (s *VaultSource) SetHttpClient(httpClient *http.Client) {
	s.HttpClient = httpClient
}

func (s *VaultSource) GetTagIds() map[string]string {
	return s.TagIds
}

func (s *VaultSource) Watch(done chan bool, group *sync.WaitGroup) {
	<-done
}
