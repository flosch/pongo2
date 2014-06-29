package pongo2

import (
	"io/ioutil"
)

func Must(tpl *Template, err error) *Template {
	if err != nil {
		panic(err)
	}
	return tpl
}

func FromString(tpl string) (*Template, error) {
	t, err := newTemplateString(tpl)
	return t, err
}

func FromFile(filename string) (*Template, error) {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	t, err := newTemplate(filename, string(buf))
	return t, err
}
