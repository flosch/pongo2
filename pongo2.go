package pongo2

import (
	"io/ioutil"
	"path/filepath"
)

func NewTemplateFromString(tpl string) (*Template, error) {
	t, err := newTemplateString(tpl)
	return t, err
}

func NewTemplateFromFile(filename string) (*Template, error) {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	t, err := newTemplate(filepath.Base(filename), string(buf))
	return t, err
}
