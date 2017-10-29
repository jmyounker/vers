package main

import (
	"testing"
	"io/ioutil"
	"os"
)

func failWhenErr(t *testing.T, err error) {
	if err != nil {
		t.Log(err)
		t.Fail()
	}
}

func failWhen(t *testing.T, x bool) {
	if x {
		t.Fail()
	}
}

func TestWriteInitFileProducesReadableFile(t *testing.T) {
	tf, err := ioutil.TempFile("", "version.json")
	failWhenErr(t, err)
	defer os.Remove(tf.Name())
	failWhenErr(t, createInitFile(tf.Name(), "default"))
	config, err := readConfig(tf.Name())
	failWhenErr(t, err)
	failWhen(t, config == nil)
}

func TestFileWithoutBranchesIsInvalid(t *testing.T) {
	tf, err := ioutil.TempFile("", "version.json")
	failWhenErr(t, err)
	defer os.Remove(tf.Name())
	c := Config{};
	failWhenErr(t, writeConfig(tf.Name(), c))
	config, err := readConfig(tf.Name())
	failWhen(t, err == nil)
	failWhen(t, config != nil)
}

func TestFileWithInvalidBranchesIsInvalid(t *testing.T) {
	tf, err := ioutil.TempFile("", "version.json")
	failWhenErr(t, err)
	defer os.Remove(tf.Name())
	c := Config{
		Branches: []BranchConfig{{
			BranchPattern: "",
			VersionTemplate: "",
		},
		},
	};
	failWhenErr(t, writeConfig(tf.Name(), c))
	config, err := readConfig(tf.Name())
	failWhen(t, err == nil)
	failWhen(t, config != nil)
}

func TestBranchConfig(t *testing.T) {
	var testBranchConfig = []struct{
		BranchPattern string
		VersionTemplate string
		ErrorValue string
	} {
		{ "", "{branch}", "branch pattern required"},
		{ ".*", "", "version template required"},
		{ "(", "{branch}", "branch pattern '(' is malformed"},
		{ ".*", "{", "version template '{' is malformed"},
	}
	for _, tc := range(testBranchConfig) {
		bc := BranchConfig{
			BranchPattern: tc.BranchPattern,
			VersionTemplate: tc.VersionTemplate,
		}
		err := checkBranchConfig(bc)
		if err == nil || err.Error() != tc.ErrorValue {
			t.Log("Wanted ", tc.ErrorValue, ", Got: ", err)
			t.Fail()
		}
	}
}

