package config_merger

import (
	"github.com/hashicorp/go-multierror"
	"github.com/stretchr/testify/assert"
	"github.com/vmware/vic/pkg/errors"
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

func TestCheckRequiredFieldsSuccessENV(t *testing.T) {
	cleanEnv(t)

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
	cleanEnv(t)

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

type testUserModel struct {
	Login    string `env:"TEST_LOGIN" required:"true"`
	Password string `env:"TEST_PASSWORD" required:"true"`
}

type testDbConnection struct {
	Port int    `env:"TEST_PORT" required:"true"`
	Host string `env:"TEST_HOST" required:"true"`
}

type testAggregateData struct {
	Comment string `env:"TEST_COMMENT" required:"true"`
	Db      testDbConnection
	User    testUserModel
}

func TestCheckRequiredFieldsSuccessENVComplexStruct(t *testing.T) {

	cleanEnv(t)

	os.Setenv("TEST_LOGIN", "TestLogin")
	os.Setenv("TEST_PORT", "8081")

	cnf := &testAggregateData{}
	envSource := &EnvSource{Variables: []string{
		"TEST_LOGIN", "TEST_PASSWORD", "TEST_PORT",
	}}

	m := NewMerger(cnf)
	m.AddSource(envSource)

	err := m.Run()

	expectredErr := &multierror.Error{}
	multierror.Append(expectredErr, errors.New("Required value Comment is empty"))
	multierror.Append(expectredErr, errors.New("Required value Host is empty"))
	multierror.Append(expectredErr, errors.New("Required value Password is empty"))

	assert.Equal(t, expectredErr, err)
}

func cleanEnv(t *testing.T) {
	err := os.Unsetenv("TEST_LOGIN")
	if err != nil {
		t.Log(err)
	}

	err = os.Unsetenv("TEST_PASSWORD")
	if err != nil {
		t.Log(err)
	}

	err = os.Unsetenv("TEST_COMMENT")
	if err != nil {
		t.Log(err)
	}

	err = os.Unsetenv("TEST_PORT")
	if err != nil {
		t.Log(err)
	}

	err = os.Unsetenv("TEST_HOST")
	if err != nil {
		t.Log(err)
	}
}
