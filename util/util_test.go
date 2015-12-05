package util

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

func TestExistsFile(t *testing.T) {
	testfile := "_testfile"

	if ExistsFile(testfile) {
		t.Errorf("file %s is exists.", testfile)
	}

	err := ioutil.WriteFile(testfile, []byte("test"), 0644)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(testfile)

	if !ExistsFile(testfile) {
		t.Errorf("file %s is not exists.", testfile)
	}
}

func TestTrim(t *testing.T) {
	expecteds := []struct {
		test     string
		expected string
	}{
		{" input ", "input"},
		{"	input	", "input"},
		{"　input　", "input"},
		{" 　input 　", "input"},
		{" 　	input test	 　", "input test"},
	}

	for _, s := range expecteds {
		if Trim(s.test) != s.expected {
			t.Errorf("mismatch (expected) '%s' != '%s'", s.test, s.expected)
		}
	}
}

func TestContainsStr(t *testing.T) {
	expecteds := []struct {
		test     string
		array    []string
		expected bool
	}{
		{"test", []string{}, false},
		{"test", []string{"hello", "world"}, false},
		{"test", []string{"hello", "world", "test"}, true},
		{"hello world", []string{"goto", "hello world", "test"}, true},
		{"hello world", []string{"hello", "world", "hello  world"}, false},
		{"hello world", []string{"hello world", "hello world"}, true},
	}

	for _, s := range expecteds {
		if ContainsStr(s.test, s.array) != s.expected {
			t.Errorf("mismatch (expected) '%s' in '%v' -> '%v'", s.test, s.array, s.expected)
		}
	}
}

func TestFileList(t *testing.T) {
	expecteds := []struct {
		path     string
		filter   string
		expected []string
	}{
		{".", `.+\.go$`, []string{"pid.go", "pid_test.go", "util.go", "util_test.go"}},
		{".", `.+\.conf$`, []string{}},
	}

	for _, s := range expecteds {
		if files, _ := FileList(s.path, s.filter); !reflect.DeepEqual(s.expected, files) {
			t.Errorf("mismatch (expected) %s/%s -> %v != %v", s.path, s.filter, s.expected, files)
		}
	}
}
