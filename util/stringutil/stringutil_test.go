package stringutil

import (
	"testing"
)

func TestGenerateRandomStr(t *testing.T) {
	got := GenerateRandomStr(10)
	if len(got) != 10 {
		t.Errorf("generateRandomStr() = %v, want length %v", got, 10)
	}
}

func TestGenerateRandomStrLong(t *testing.T) {
	length := 10000
	got := GenerateRandomStr(length)
	if len(got) != length {
		t.Errorf("generateRandomStr() = %v, want length %v", got, length)
	}
}

func TestGenerateRandomStrEmpty(t *testing.T) {
	got := GenerateRandomStr(0)
	if len(got) != 0 {
		t.Errorf("generateRandomStr() = %v, want length %v", got, 0)
	}
}

func TestStringSliceToJsonArray(t *testing.T) {
	got := StringSliceToJsonArray([]string{"a", "b", "c"})
	if got != "[a,b,c]" {
		t.Errorf("stringSliceToJsonArray() = %v, want %v", got, "[a,b,c]")
	}

	testStrSlice := []string{}
	got = StringSliceToJsonArray(testStrSlice)
	if got != "[]" {
		t.Errorf("stringSliceToJsonArray() = %v, want %v", got, "[]")
	}
}
