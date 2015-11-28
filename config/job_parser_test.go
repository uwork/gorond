package config

import (
	"testing"
)

// コメント正規表現のテスト
func TestCommentRegex(t *testing.T) {

	expecteds := []struct {
		test     string
		expected string
	}{
		{"hogehoge # comment", "hogehoge "},
		{"hogehoge #", "hogehoge "},
		{"hoge#hoge", "hoge"},
		{"hoge = '#hash' #comment", "hoge = '#hash' "},
		{`hoge = "#hash" #comment`, `hoge = "#hash" `},
		{"hoge = '#hash' #co'mment", "hoge = '#hash' "},
		{"hoge #hash #comment", "hoge "},
		{"hoge #hash #com'men't #commt", "hoge "},
		{"#com'men't #commt", ""},
		{"   ##comment", "   "},
	}

	for _, s := range expecteds {
		line := commentRegex.ReplaceAllString(s.test, "$1")
		if s.expected != line {
			t.Errorf("%s => (expected) (%s) != (%s)", s.test, s.expected, line)
		}
	}

}

// パースのテスト
func TestParseJobLine(t *testing.T) {
	expecteds := []struct {
		test     string
		expected Job
	}{
		{"0 * * * * ? root /bin/echo test; `id`", Job{"", "0 * * * * ?", "root", "/bin/echo test; `id`", 0, WAITING, []*Job{}, nil}},
		{"@yearly  user   /bin/echo foobar", Job{"", "@yearly", "user", "/bin/echo foobar", 0, WAITING, []*Job{}, nil}},
		{"@everly 1h30m  root /bin/echo hogefuga", Job{"", "@everly 1h30m", "root", "/bin/echo hogefuga", 0, WAITING, []*Job{}, nil}},
		{"  - root date", Job{"", "", "root", "date", 2, WAITING, []*Job{}, nil}},
		{"    - root date", Job{"", "", "root", "date", 4, WAITING, []*Job{}, nil}},
		{"@yearly     /bin/echo error", Job{}},
		{"		- root /bin/echo error", Job{}},
	}

	for _, s := range expecteds {
		actual, err := parseJobLine(s.test)
		if s.expected.Command == "" && err == nil {
			t.Errorf("Parse   : %s => (expected)", s.test)
		}
		if actual.Schedule != s.expected.Schedule {
			t.Errorf("Schedule: %s => (expected) '%s' != '%s'", s.test, actual.Schedule, s.expected.Schedule)
		}
		if actual.Command != s.expected.Command {
			t.Errorf("Command : %s => (expected) '%s' != '%s'", s.test, actual.Command, s.expected.Command)
		}
		if actual.User != s.expected.User {
			t.Errorf("User    : %s => (expected) '%s' != '%s'", s.test, actual.User, s.expected.User)
		}
		if actual.Indent != s.expected.Indent {
			t.Errorf("Indent  : %s => (expected) '%d' != '%d'", s.test, actual.Indent, s.expected.Indent)
		}
		if actual.Status != s.expected.Status {
			t.Errorf("Status  : %s => (expected) '%d' != '%d'", s.test, actual.Status, s.expected.Status)
		}
	}

}

// ツリー構造のテスト
func TestJobTree(t *testing.T) {

	s := struct {
		test     string
		expected Job
	}{
		`
# test
0 * * * * ? root /bin/echo test
            - root /bin/echo test2
            - root /bin/echo test3
              - root /bin/echo test4
`, Job{"", "0 * * * * ?", "root", "/bin/echo test", 0, WAITING, []*Job{
			&Job{"", "", "root", "/bin/echo test2", 12, WAITING, []*Job{}, nil},
			&Job{"", "", "root", "/bin/echo test3", 12, WAITING, []*Job{
				&Job{"", "", "root", "/bin/echo test4", 14, WAITING, []*Job{}, nil},
			}, nil},
		}, nil},
	}

	actual, err := parseJobConfig(s.test)
	if err != nil {
		t.Errorf("parse error: %v", err)
		return
	}
	if len(actual[0].Childs) != len(s.expected.Childs) {
		t.Errorf("child count missmatch '%s' (expected) %d != %d", actual[0].Line, len(s.expected.Childs), len(actual[0].Childs))
	}
	if len(actual[1].Childs) != 0 {
		t.Errorf("child count missmatch '%s' (expected) %d != %d", actual[1].Line, len(s.expected.Childs[0].Childs), len(actual[1].Childs))
	}
	if len(actual[2].Childs) != 1 {
		t.Errorf("child count missmatch '%s' (expected) %d != %d", actual[2].Line, len(s.expected.Childs[1].Childs), len(actual[2].Childs))
	}
	if len(actual[3].Childs) != 0 {
		t.Errorf("child count missmatch '%s' (expected) %d != %d", actual[2].Line, len(s.expected.Childs[1].Childs[0].Childs), len(actual[3].Childs))
	}
}

// ツリー構造のエラーテスト
func TestJobTreeError(t *testing.T) {

	test := `
0 * * * * ? root /bin/echo test
            - root /bin/echo test2
          - root /bin/echo test3
`
	_, err := parseJobConfig(test)
	if err == nil {
		t.Errorf("valid config '%s' (expected) error", test)
	}
}
