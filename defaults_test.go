package config_merger

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

type testAggregateDataDefaults struct {
	Comment1 string `env:"TEST_COMMENT1" default:"default comment 1"`
	Comment2 string `env:"TEST_COMMENT2" default:"default comment 2"`
	Comment3 string `env:"TEST_COMMENT3" default:"default comment 3"`
	Count    int    `env:"TEST_COUNT" default:"10"`
}

type testNestedComment struct {
	Comment4 string `env:"TEST_COMMENT4" default:"default comment 4"`
	Comment5 string `env:"TEST_COMMENT5" default:"default comment 5"`
}

type testNested struct {
	Value1   string `env:"TEST_VALUE1" required:"true"`
	Comments testNestedComment
}

func TestDefaultsNested_setDefaults(t *testing.T) {

	cleanEnv(t)

	os.Setenv("TEST_VALUE1", "TestValue1")
	os.Setenv("TEST_COMMENT4", "TestComment4FromEnv")

	cnf := testNested{}
	envSource := &EnvSource{Variables: []string{
		"TEST_VALUE1", "TEST_COMMENT4",
	}}

	m := NewMerger(&cnf)
	m.AddSource(envSource)

	err := m.Run()

	assert.NoError(t, err)
	assert.Equal(t, "TestValue1", cnf.Value1)
	assert.Equal(t, "TestComment4FromEnv", cnf.Comments.Comment4)
	assert.Equal(t, "default comment 5", cnf.Comments.Comment5)
}

func TestDefaults_setDefaults(t *testing.T) {

	cleanEnv(t)

	cnf := testAggregateDataDefaults{}
	envSource := &EnvSource{Variables: []string{
		"TEST_COMMENT3", "TEST_COUNT",
	}}

	m := NewMerger(&cnf)
	m.AddSource(envSource)

	err := m.Run()

	assert.NoError(t, err)
	assert.Equal(t, "default comment 1", cnf.Comment1)
	assert.Equal(t, "default comment 2", cnf.Comment2)
	assert.Equal(t, "default comment 3", cnf.Comment3)
	assert.Equal(t, 10, cnf.Count)
}
