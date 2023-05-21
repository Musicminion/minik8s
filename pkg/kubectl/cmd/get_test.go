package cmd

import "testing"

func TestParseNameAndNamespace(t *testing.T) {
	namespace, name, err := parseNameAndNamespace("default/pod1")
	if err != nil {
		t.Error(err)
	}

	if namespace != "default" {
		t.Error("namespace should be default")
	}

	if name != "pod1" {
		t.Error("name should be pod1")
	}
}
