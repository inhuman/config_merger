package config_merger

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
	"net/http"
	"testing"
	"time"
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
		Token:   "dummy_token",
		Prefix:  "secret/test/config_merger",
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
		Token:   "dummy_token",
		Prefix:  "secret/test/config_merger",
	}
	cnf := &Cnf{}

	client := &http.Client{Transport: &http.Transport{}, Timeout: time.Microsecond}
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
		Token:   "dummy_token",
		Prefix:  "secret/test/config_merger",
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

type NestedDobleConf struct {
	NestedPassword struct {
		Pass1 string `json:"pass_1" vault:"service1/db1:password"`
		Pass2 string `json:"pass_2" vault:"service2/db2:password"`
	} `json:"nested_password"`
}

func TestVaultSource_LoadNestedConfigDoubleFields(t *testing.T) {

	defer gock.Off()

	gock.New("http://vault.example.local").
		Get("/v1/secret/service1/db1").
		Reply(200).
		BodyString(`{
    "request_id": "fde655e4-d755-6cad-10e5-c0ed790b66ed",
    "lease_id": "",
    "renewable": false,
    "lease_duration": 2764800,
    "data": {
		"password": "db_nested_password1"
	},
    "wrap_info": null,
    "warnings": null,
    "auth": null
}`)

	gock.New("http://vault.example.local").
		Get("/v1/secret/service2/db2").
		Reply(200).
		BodyString(`{
    "request_id": "fde655e4-d755-6cad-10e5-c0ed790b66ed",
    "lease_id": "",
    "renewable": false,
    "lease_duration": 2764800,
    "data": {
		"password": "db_nested_password2"
	},
    "wrap_info": null,
    "warnings": null,
    "auth": null
}`)

	vaultSource := &VaultSource{
		Address: "http://vault.example.local",
		Token:   "dummy_token",
		Prefix:  "secret/",
	}
	cnf := &NestedDobleConf{}

	client := &http.Client{Transport: &http.Transport{}, Timeout: time.Microsecond}
	gock.InterceptClient(client)

	vaultSource.SetHttpClient(client)
	vaultSource.SetTargetStruct(cnf)

	err := vaultSource.Load()
	assert.NoError(t, err)
	assert.Equal(t, "db_nested_password1", cnf.NestedPassword.Pass1)
	assert.Equal(t, "db_nested_password2", cnf.NestedPassword.Pass2)
}

type NestedConfDuplicate struct {
	NestedPassword struct {
		Pass1 string `json:"pass_1" vault:"pass_1"`
		Pass2 string `json:"pass_2" vault:"pass_1"`
	} `json:"nested_password"`
}

func TestMerger_DuplicateVaultTags(t *testing.T) {

	cnf := &NestedConfDuplicate{}
	var panicMessage = "duplicate vault tags found: pass_1"

	defer func() {
		if r := recover(); r != nil {
			if r != panicMessage {

				fmt.Println(r)

				t.Error("Panic message expected '" + panicMessage + "', got '" + r.(string) + "'")
			}
		}
	}()
	NewMerger(cnf)
	t.Error("The code did not panic")
}

func TestMerger_VaultSourceParsePrefixName(t *testing.T) {

	vaultSource := &VaultSource{}
	var prefix, name string
	var err error

	t.Log("Test success parse (main prefix + tag prefix + field)")
	vaultSource.Prefix = "secret/"
	prefix, name, err = vaultSource.getPrefixAndName("test/config_merger:password")
	assert.NoError(t, err)
	assert.Equal(t, "test/config_merger", prefix)
	assert.Equal(t, "password", name)

	t.Log("Test success parse (main prefix + field)")
	vaultSource.Prefix = "secret/test/config_merger"
	prefix, name, err = vaultSource.getPrefixAndName("password")
	assert.NoError(t, err)
	assert.Equal(t, "", prefix)
	assert.Equal(t, "password", name)

	t.Log("Test fail parse (missing prefix)")
	vaultSource.Prefix = ""
	prefix, name, err = vaultSource.getPrefixAndName("password")
	assert.Equal(t, "Unconditional tag `password` - missing prefix", err.Error())
	assert.Equal(t, "", prefix)
	assert.Equal(t, "", name)

	t.Log("Test fail parse (missing field)")
	vaultSource.Prefix = ""
	prefix, name, err = vaultSource.getPrefixAndName("secret/test/config_merger")
	assert.Equal(t, "Unconditional tag `secret/test/config_merger` - missing field name", err.Error())
	assert.Equal(t, "", prefix)
	assert.Equal(t, "", name)

	t.Log("Test fail parse (unconditional tag)")
	vaultSource.Prefix = ""
	prefix, name, err = vaultSource.getPrefixAndName("secret/test/config_merger:password:password")
	assert.Equal(t, "Unconditional tag `secret/test/config_merger:password:password`", err.Error())
	assert.Equal(t, "", prefix)
	assert.Equal(t, "", name)
}

type CnfTag struct {
	Message string `json:"message" tagId:"cnf_tag_message"`
}

func TestVaultSource_TagIdsLoadSuccess(t *testing.T) {
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
		Token:   "dummy_token",
		Prefix:  "secret/test/config_merger",
	}
	vaultSource.TagIds = map[string]string{
		"cnf_tag_message": "message",
	}

	cnf := &CnfTag{}

	client := &http.Client{Transport: &http.Transport{}, Timeout: time.Microsecond}
	gock.InterceptClient(client)

	vaultSource.SetHttpClient(client)
	vaultSource.SetTargetStruct(cnf)

	err := vaultSource.Load()
	assert.NoError(t, err)
	assert.Equal(t, "from vault", cnf.Message)
}
