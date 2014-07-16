package pongo2

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

var admin_list = []string{"user2"}

var time1 = time.Date(2014, 06, 10, 15, 30, 15, 0, time.UTC)
var time2 = time.Date(2011, 03, 21, 8, 37, 56, 12, time.UTC)

type post struct {
	Text    string
	Created time.Time
}

type user struct {
	Name      string
	Validated bool
}

type comment struct {
	Author *user
	Date   time.Time
	Text   string
}

func is_admin(in *Value) bool {
	u, worked := in.Interface().(*user)
	if !worked {
		return false
	}
	for _, a := range admin_list {
		if a == u.Name {
			return true
		}
	}
	return false
}

func (u *user) Is_admin() *Value {
	return AsValue(is_admin(AsValue(u)))
}

func (u *user) Is_admin2() bool {
	return is_admin(AsValue(u))
}

func (p *post) String() string {
	return ":-)"
}

var tplContext = Context{
	"number": 11,
	"simple": map[string]interface{}{
		"number":        42,
		"name":          "john doe",
		"included_file": "INCLUDES.helper",
		"nil":           nil,
		"uint":          uint(8),
		"float":         float64(3.1415),
		"str":           "string",
		"bool_true":     true,
		"bool_false":    false,
		"newline_text": `this is a text
with a new line in it`,
		"intmap": map[int]string{
			1: "one",
			2: "two",
			5: "five",
		},
		"func_add": func(a, b *Value) int {
			return a.Integer() + b.Integer()
		},
	},
	"complex": map[string]interface{}{
		"is_admin": is_admin,
		"post": post{
			Text:    "<h2>Hello!</h2><p>Welcome to my new blog page. I'm using pongo2 which supports {{ variables }} and {% tags %}.</p>",
			Created: time2,
		},
		"comments": []*comment{
			&comment{
				Author: &user{
					Name:      "user1",
					Validated: true,
				},
				Date: time1,
				Text: "\"pongo2 is nice!\"",
			},
			&comment{
				Author: &user{
					Name:      "user2",
					Validated: true,
				},
				Date: time2,
				Text: "comment2 with <script>unsafe</script> tags in it",
			},
			&comment{
				Author: &user{
					Name:      "user3",
					Validated: false,
				},
				Date: time1,
				Text: "<b>hello!</b> there",
			},
		},
	},
}

func TestTemplates(t *testing.T) {
	SetDebug(true) // activate pongo2's debugging output

	matches, err := filepath.Glob("./template_tests/*.tpl")
	if err != nil {
		t.Fatal(err)
	}
	for idx, match := range matches {
		t.Logf("[Template %3d] Testing '%s'", idx+1, match)
		tpl, err := FromFile(match)
		if err != nil {
			t.Fatalf("Error on FromFile('%s'): %s", match, err.Error())
		}
		test_filename := fmt.Sprintf("%s.out", match)
		test_out, err := ioutil.ReadFile(test_filename)
		if err != nil {
			t.Fatalf("Error on ReadFile('%s'): %s", test_filename, err.Error())
		}
		tpl_out, err := tpl.Execute(tplContext)
		if err != nil {
			t.Fatalf("Error on Execute('%s'): %s", match, err.Error())
		}
		if string(test_out) != tpl_out {
			t.Logf("Template (rendered) '%s': '%s'", match, tpl_out)
			err_filename := "tpl-error.out"
			ioutil.WriteFile(err_filename, []byte(tpl_out), 0700)
			t.Logf("get a complete diff with command: 'diff -ya %s %s'", test_filename, err_filename)
			t.Fatalf("Failed: test_out != tpl_out for %s", match)
		}
	}
}

func TestExecutionErrors(t *testing.T) {
	SetDebug(true) // activate pongo2's debugging output

	matches, err := filepath.Glob("./template_tests/*-execution.err")
	if err != nil {
		t.Fatal(err)
	}
	for idx, match := range matches {
		t.Logf("[Errors %3d] Testing '%s'", idx+1, match)

		test_data, err := ioutil.ReadFile(match)
		tests := strings.Split(string(test_data), "\n")

		check_filename := fmt.Sprintf("%s.out", match)
		check_data, err := ioutil.ReadFile(check_filename)
		if err != nil {
			t.Fatalf("Error on ReadFile('%s'): %s", check_filename, err.Error())
		}
		checks := strings.Split(string(check_data), "\n")

		if len(checks) != len(tests) {
			t.Fatal("Template lines != Checks lines")
		}

		for idx, test := range tests {
			if strings.TrimSpace(test) == "" {
				continue
			}
			if strings.TrimSpace(checks[idx]) == "" {
				t.Fatalf("[%s Line %d] Check is empty (must contain an regular expression).",
					match, idx+1)
			}

			tpl, err := FromString(test)
			if err != nil {
				t.Fatalf("Error on FromString('%s'): %s", test, err.Error())
			}

			_, err = tpl.Execute(tplContext)
			if err == nil {
				t.Fatalf("[%s Line %d] Expected error for (got none): %s",
					match, idx+1, tests[idx])
			}
			re := regexp.MustCompile(fmt.Sprintf("^%s$", checks[idx]))
			if !re.MatchString(err.Error()) {
				t.Fatalf("[%s Line %d] Error for '%s' (err = '%s') does not match the (regexp-)check: %s",
					match, idx+1, test, err.Error(), checks[idx])
			}
		}
	}
}

func TestCompilationErrors(t *testing.T) {
	SetDebug(true) // activate pongo2's debugging output

	matches, err := filepath.Glob("./template_tests/*-compilation.err")
	if err != nil {
		t.Fatal(err)
	}
	for idx, match := range matches {
		t.Logf("[Errors %3d] Testing '%s'", idx+1, match)

		test_data, err := ioutil.ReadFile(match)
		tests := strings.Split(string(test_data), "\n")

		check_filename := fmt.Sprintf("%s.out", match)
		check_data, err := ioutil.ReadFile(check_filename)
		if err != nil {
			t.Fatalf("Error on ReadFile('%s'): %s", check_filename, err.Error())
		}
		checks := strings.Split(string(check_data), "\n")

		if len(checks) != len(tests) {
			t.Fatal("Template lines != Checks lines")
		}

		for idx, test := range tests {
			if strings.TrimSpace(test) == "" {
				continue
			}
			if strings.TrimSpace(checks[idx]) == "" {
				t.Fatalf("[%s Line %d] Check is empty (must contain an regular expression).",
					match, idx+1)
			}

			_, err = FromString(test)
			if err == nil {
				t.Fatalf("[%s | Line %d] Expected error for (got none): %s", match, idx+1, tests[idx])
			}
			re := regexp.MustCompile(fmt.Sprintf("^%s$", checks[idx]))
			if !re.MatchString(err.Error()) {
				t.Fatalf("[%s | Line %d] Error for '%s' (err = '%s') does not match the (regexp-)check: %s",
					match, idx+1, test, err.Error(), checks[idx])
			}
		}
	}
}

func BenchmarkExecuteComplex(b *testing.B) {
	tpl, err := FromFile("template_tests/complex.tpl")
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err = tpl.Execute(tplContext)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCompileAndExecuteComplex(b *testing.B) {
	buf, err := ioutil.ReadFile("template_tests/complex.tpl")
	if err != nil {
		b.Fatal(err)
	}
	preloadedTpl := string(buf)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tpl, err := FromString(preloadedTpl)
		if err != nil {
			b.Fatal(err)
		}

		_, err = tpl.Execute(tplContext)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParallelExecuteComplex(b *testing.B) {
	tpl, err := FromFile("template_tests/complex.tpl")
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err = tpl.Execute(tplContext)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
