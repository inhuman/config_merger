package config_merger

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

type TestStructJSON struct {
	RequiredField string `json:"required_field" required:"true"`
	Empty         string `json:"empty"`
}

func TestCheckRequiredFieldsSuccessJSON(t *testing.T) {

	fh := createFileForTest(t, `{"required_field": "from json"}`)
	path := fh.Name()
	defer func() {
		fh.Close()
		os.Remove(path)
	}()

	jsonSource := &JsonSource{Path: path}
	cnf := &TestStructJSON{}

	m := NewMerger(cnf)
	m.AddSource(jsonSource)
	err := m.Run()

	assert.NoError(t, err)
	assert.Equal(t, &TestStructJSON{RequiredField: "from json"}, cnf)

}

func TestCheckRequiredFieldsFailJSON(t *testing.T) {

	fh := createFileForTest(t, `{"message": "from json"}`)
	path := fh.Name()
	defer func() {
		fh.Close()
		os.Remove(path)
	}()

	jsonSource := &JsonSource{Path: path}

	cnf := &TestStructJSON{}

	m := NewMerger(cnf)
	m.AddSource(jsonSource)

	err := m.Run()

	assert.Error(t, err)
}

type TestStructENV struct {
	Login    string `env:"TEST_LOGIN" required:"true"`
	Password string `env:"TEST_PASSWORD" required:"true"`
	Comment  string `env:"TEST_COMMENT"`
	Port     int    `env:"TEST_PORT" required:"true"`
}

func cleanEnv() {
	os.Unsetenv("TEST_LOGIN")
	os.Unsetenv("TEST_PASSWORD")
	os.Unsetenv("TEST_COMMENT")
	os.Unsetenv("TEST_PORT")
}

func TestCheckRequiredFieldsSuccessENV(t *testing.T) {
	cleanEnv()

	os.Setenv("TEST_LOGIN", "TestLogin")
	os.Setenv("TEST_PASSWORD", "te$t_pasw00rd")
	os.Setenv("TEST_COMMENT", "Test user comment")
	os.Setenv("TEST_PORT", "8081")

	cnf := &TestStructENV{}
	envSource := &EnvSource{Variables: []string{
		"TEST_LOGIN", "TEST_PASSWORD", "TEST_PORT", "TEST_COMMENT",
	}}

	m := NewMerger(cnf)
	m.AddSource(envSource)

	err := m.Run()
	assert.NoError(t, err)

	assert.Equal(t, "TestLogin", cnf.Login)
	assert.Equal(t, "te$t_pasw00rd", cnf.Password)
	assert.Equal(t, "Test user comment", cnf.Comment)
	assert.Equal(t, 8081, cnf.Port)

}

func TestCheckRequiredFieldsFailENV(t *testing.T) {
	cleanEnv()

	os.Setenv("TEST_LOGIN", "TestLogin")
	os.Setenv("TEST_PORT", "8081")

	cnf := &TestStructENV{}
	envSource := &EnvSource{Variables: []string{
		"TEST_LOGIN", "TEST_PASSWORD", "TEST_PORT",
	}}

	m := NewMerger(cnf)
	m.AddSource(envSource)

	err := m.Run()
	assert.Error(t, err, "* Required value Password is empty")
}
