package pongo2

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"testing"
)

type DummyLoader struct{}

func (l *DummyLoader) Abs(base, name string) string {
	return filepath.Join(filepath.Dir(base), name)
}

func (l *DummyLoader) Get(path string) (io.Reader, error) {
	return nil, errors.New("dummy not found")
}

func FuzzBuiltinFilters(f *testing.F) {
	f.Add("foobar", "123")
	f.Add("foobar", `123,456`)
	f.Add("foobar", `123,456,"789"`)
	f.Add("foobar", `"test","test"`)
	f.Add("foobar", `123,"test"`)
	f.Add("foobar", "")
	f.Add("123", "foobar")

	f.Fuzz(func(t *testing.T, value, filterArg string) {
		ts := NewSet("fuzz-test", &DummyLoader{})
		for name := range filters {
			tpl, err := ts.FromString(fmt.Sprintf("{{ %v|%v:%v }}", value, name, filterArg))
			if tpl != nil && err != nil {
				t.Errorf("filter=%q value=%q, filterArg=%q, err=%v", name, value, filterArg, err)
			}
			if err == nil {
				tpl.Execute(nil)
			}
		}
	})
}
