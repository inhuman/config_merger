package config_merger

import (
	"testing"
	"gopkg.in/h2non/gock.v1"
	"net/http"
	"github.com/stretchr/testify/assert"
	"time"
	"fmt"
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

	client := &http.Client{Transport: &http.Transport{}, Timeout: time.Microsecond}
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

	client := &http.Client{Transport: &http.Transport{},  Timeout: time.Microsecond}
	gock.InterceptClient(client)

	vaultSource.SetHttpClient(client)
	vaultSource.SetTargetStruct(cnf)
	err := vaultSource.Load()
	assert.Error(t, err)
}


func TestVaultSource_LoadNestedConfig(t *testing.T) {

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
		"pass_1": "nested_password1",
    	"pass_2": "nested_password2"
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
	cnf := &NestedConf{}

	client := &http.Client{Transport: &http.Transport{}, Timeout: time.Microsecond}
	gock.InterceptClient(client)

	vaultSource.SetHttpClient(client)
	vaultSource.SetTargetStruct(cnf)

	err := vaultSource.Load()
	assert.NoError(t, err)
	assert.Equal(t, "nested_password1", cnf.NestedPassword.Pass1)
	assert.Equal(t, "nested_password2", cnf.NestedPassword.Pass2)
}


type NestedConfDuplicate struct {
	NestedPassword struct{
		Pass1 string `json:"pass_1" vault:"pass_1"`
		Pass2 string `json:"pass_2" vault:"pass_1"`
	} `json:"nested_password"`
}


func TestMerger_DuplicateVaultTags(t *testing.T) {

	cnf := &NestedConfDuplicate{}
	var panicMessage = "duplicate vault tags found: pass_1"

	defer func() {
		if r := recover(); r != nil {
			if r != panicMessage{

				fmt.Println(r)

				t.Error("Panic message expected '" + panicMessage + "', got '" + r.(string) + "'")
			}
		}
	}()
	NewMerger(cnf)
	t.Error("The code did not panic")
}

