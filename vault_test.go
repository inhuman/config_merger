package config_merger

import (
	"testing"
	"gopkg.in/h2non/gock.v1"
	"net/http"
	"github.com/stretchr/testify/assert"
)

func TestVaultSource_LoadSuccess(t *testing.T) {

	defer gock.Off()

	gock.New("http://vault.example.local").
		Get("/v1/secret/test/config_merger").
		Reply(200).
		BodyString(`{
    "request_id": "fde655e4-d755-6cad-10e5-c0ed790b66ed",
    "lease_id": "",
    "renewable": false,
    "lease_duration": 2764800,
    "data": {
        "message": "from vault"
    },
    "wrap_info": null,
    "warnings": null,
    "auth": null
}`)

	vaultSource := &VaultSource{
		Address: "http://vault.example.local",
		Token: "dummy_token",
		Prefix: "secret/test/config_merger",
	}
	cnf := &Cnf{}

	client := &http.Client{Transport: &http.Transport{}}
	gock.InterceptClient(client)

	vaultSource.SetHttpClient(client)
	vaultSource.SetTargetStruct(cnf)


	err := vaultSource.Load()
	assert.NoError(t, err)
	assert.Equal(t, "from vault", cnf.Message)
}

func TestVaultSource_LoadWrongAddr(t *testing.T) {

	vaultSource := &VaultSource{
		Address: "http://vault.example.local",
		Token: "dummy_token",
		Prefix: "secret/test/config_merger",
	}
	cnf := &Cnf{}

	client := &http.Client{Transport: &http.Transport{}}
	gock.InterceptClient(client)

	vaultSource.SetHttpClient(client)
	vaultSource.SetTargetStruct(cnf)
	err := vaultSource.Load()
	assert.Error(t, err)
}