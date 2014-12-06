package pongo2

import (
	"bytes"
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

func is_admin(u *user) bool {
	for _, a := range admin_list {
		if a == u.Name {
			return true
		}
	}
	return false
}

func (u *user) Is_admin() *Value {
	return AsValue(is_admin(u))
}

func (u *user) Is_admin2() bool {
	return is_admin(u)
}

func (p *post) String() string {
	return ":-)"
}

/*
 * Start setup sandbox
 */

type tagSandboxDemoTag struct {
}

func (node *tagSandboxDemoTag) Execute(ctx *ExecutionContext, buffer *bytes.Buffer) *Error {
	buffer.WriteString("hello")
	return nil
}

func tagSandboxDemoTagParser(doc *Parser, start *Token, arguments *Parser) (INodeTag, *Error) {
	return &tagSandboxDemoTag{}, nil
}

func BannedFilterFn(in *Value, params *Value) (*Value, *Error) {
	return in, nil
}

func init() {
	DefaultSet.Debug = true

	RegisterFilter("banned_filter", BannedFilterFn)
	RegisterFilter("unbanned_filter", BannedFilterFn)
	RegisterTag("banned_tag", tagSandboxDemoTagParser)
	RegisterTag("unbanned_tag", tagSandboxDemoTagParser)

	DefaultSet.BanFilter("banned_filter")
	DefaultSet.BanTag("banned_tag")

	// Allow different kind of levels inside template_tests/
	abs_path, err := filepath.Abs("./template_tests/*")
	if err != nil {
		panic(err)
	}
	DefaultSet.SandboxDirectories = append(DefaultSet.SandboxDirectories, abs_path)

	abs_path, err = filepath.Abs("./template_tests/*/*")
	if err != nil {
		panic(err)
	}
	DefaultSet.SandboxDirectories = append(DefaultSet.SandboxDirectories, abs_path)

	abs_path, err = filepath.Abs("./template_tests/*/*/*")
	if err != nil {
		panic(err)
	}
	DefaultSet.SandboxDirectories = append(DefaultSet.SandboxDirectories, abs_path)

	// Allow pongo2 temp files
	DefaultSet.SandboxDirectories = append(DefaultSet.SandboxDirectories, "/tmp/pongo2_*")

	f, err := ioutil.TempFile("/tmp/", "pongo2_")
	if err != nil {
		panic("cannot write to /tmp/")
	}
	f.Write([]byte("Hello from pongo2"))
	DefaultSet.Globals["temp_file"] = f.Name()
}

/*
 * End setup sandbox
 */

var tplContext = Context{
	"number": 11,
	"simple": map[string]interface{}{
		"number":        42,
		"name":          "john doe",
		"included_file": "INCLUDES.helper",
		"included_file_not_exists": "INCLUDES.helper.not_exists",
		"nil":           nil,
		"uint":          uint(8),
		"float":         float64(3.1415),
		"str":           "string",
		"chinese_hello_world": "你好世界",
		"bool_true":           true,
		"bool_false":          false,
		"newline_text": `this is a text
with a new line in it`,
		"long_text": `This is a simple text.

This too, as a paragraph.
Right?

Yep!`,
		"escape_js_test":     `escape sequences \r\n\'\" special chars "?!=$<>`,
		"one_item_list":      []int{99},
		"multiple_item_list": []int{1, 1, 2, 3, 5, 8, 13, 21, 34, 55},
		"misc_list":          []interface{}{"Hello", 99, 3.14, "good"},
		"escape_text":        "This is \\a Test. \"Yep\". 'Yep'.",
		"xss":                "<script>alert(\"uh oh\");</script>",
		"intmap": map[int]string{
			1: "one",
			2: "two",
			5: "five",
		},
		"func_add": func(a, b int) int {
			return a + b
		},
		"func_add_iface": func(a, b interface{}) interface{} {
			return a.(int) + b.(int)
		},
		"func_variadic": func(msg string, args ...interface{}) string {
			return fmt.Sprintf(msg, args...)
		},
		"func_variadic_sum_int": func(args ...int) int {
			// Create a sum
			s := 0
			for _, i := range args {
				s += i
			}
			return s
		},
		"func_variadic_sum_int2": func(args ...*Value) *Value {
			// Create a sum
			s := 0
			for _, i := range args {
				s += i.Integer()
			}
			return AsValue(s)
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
		"comments2": []*comment{
			&comment{
				Author: &user{
					Name:      "user1",
					Validated: true,
				},
				Date: time2,
				Text: "\"pongo2 is nice!\"",
			},
			&comment{
				Author: &user{
					Name:      "user1",
					Validated: true,
				},
				Date: time1,
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
	debug = true

	// Add a global to the default set
	Globals["this_is_a_global_variable"] = "this is a global text"

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
		test_out, rerr := ioutil.ReadFile(test_filename)
		if rerr != nil {
			t.Fatalf("Error on ReadFile('%s'): %s", test_filename, rerr.Error())
		}
		tpl_out, err := tpl.ExecuteBytes(tplContext)
		if err != nil {
			t.Fatalf("Error on Execute('%s'): %s", match, err.Error())
		}
		if bytes.Compare(test_out, tpl_out) != 0 {
			t.Logf("Template (rendered) '%s': '%s'", match, tpl_out)
			err_filename := filepath.Base(fmt.Sprintf("%s.error", match))
			err := ioutil.WriteFile(err_filename, []byte(tpl_out), 0600)
			if err != nil {
				t.Fatalf(err.Error())
			}
			t.Logf("get a complete diff with command: 'diff -ya %s %s'", test_filename, err_filename)
			t.Errorf("Failed: test_out != tpl_out for %s", match)
		}
	}
}

func TestExecutionErrors(t *testing.T) {
	debug = true

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

			_, err = tpl.ExecuteBytes(tplContext)
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
	debug = true

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

func TestBaseDirectory(t *testing.T) {
	mustStr := "Hello from template_tests/base_dir_test/"

	s := NewSet("test set with base directory")
	s.Globals["base_directory"] = "template_tests/base_dir_test/"
	if err := s.SetBaseDirectory(s.Globals["base_directory"].(string)); err != nil {
		t.Fatal(err)
	}

	matches, err := filepath.Glob("./template_tests/base_dir_test/subdir/*")
	if err != nil {
		t.Fatal(err)
	}
	for _, match := range matches {
		match = strings.Replace(match, "template_tests/base_dir_test/", "", -1)

		tpl, err := s.FromFile(match)
		if err != nil {
			t.Fatal(err)
		}
		out, err := tpl.Execute(nil)
		if err != nil {
			t.Fatal(err)
		}
		if out != mustStr {
			t.Errorf("%s: out ('%s') != mustStr ('%s')", match, out, mustStr)
		}
	}
}

func BenchmarkCache(b *testing.B) {
	cache_set := NewSet("cache set")
	for i := 0; i < b.N; i++ {
		tpl, err := cache_set.FromCache("template_tests/complex.tpl")
		if err != nil {
			b.Fatal(err)
		}
		_, err = tpl.ExecuteBytes(tplContext)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCacheDebugOn(b *testing.B) {
	cache_debug_set := NewSet("cache set")
	cache_debug_set.Debug = true
	for i := 0; i < b.N; i++ {
		tpl, err := cache_debug_set.FromFile("template_tests/complex.tpl")
		if err != nil {
			b.Fatal(err)
		}
		_, err = tpl.ExecuteBytes(tplContext)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkExecuteComplexWithSandboxActive(b *testing.B) {
	tpl, err := FromFile("template_tests/complex.tpl")
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err = tpl.ExecuteBytes(tplContext)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCompileAndExecuteComplexWithSandboxActive(b *testing.B) {
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

		_, err = tpl.ExecuteBytes(tplContext)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParallelExecuteComplexWithSandboxActive(b *testing.B) {
	tpl, err := FromFile("template_tests/complex.tpl")
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := tpl.ExecuteBytes(tplContext)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkExecuteComplexWithoutSandbox(b *testing.B) {
	s := NewSet("set without sandbox")
	tpl, err := s.FromFile("template_tests/complex.tpl")
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err = tpl.ExecuteBytes(tplContext)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCompileAndExecuteComplexWithoutSandbox(b *testing.B) {
	buf, err := ioutil.ReadFile("template_tests/complex.tpl")
	if err != nil {
		b.Fatal(err)
	}
	preloadedTpl := string(buf)

	s := NewSet("set without sandbox")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tpl, err := s.FromString(preloadedTpl)
		if err != nil {
			b.Fatal(err)
		}

		_, err = tpl.ExecuteBytes(tplContext)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParallelExecuteComplexWithoutSandbox(b *testing.B) {
	s := NewSet("set without sandbox")
	tpl, err := s.FromFile("template_tests/complex.tpl")
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := tpl.ExecuteBytes(tplContext)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
