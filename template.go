package pongo2

import (
	"net/http"
)

type Template struct {
	// Input
	name string
	tpl  string

	// Calculation
	tokens []*Token
	parser *Parser

	// first come, first serve (it's important to not override existing entries in here)
	level  int
	parent *Template
	child  *Template
	blocks map[string]*NodeWrapper

	// Output
	root *NodeDocument
}

func newTemplateString(tpl string) (*Template, error) {
	return newTemplate("<string>", tpl)
}

func newTemplate(name, tpl string) (*Template, error) {
	// Create the template
	t := &Template{
		name:   name,
		tpl:    tpl,
		blocks: make(map[string]*NodeWrapper),
	}

	// Tokenize it
	tokens, err := lex(name, tpl)
	if err != nil {
		return nil, err
	}
	t.tokens = tokens

	// For debugging purposes, show all tokens:
	/*for i, t := range tokens {
		fmt.Printf("%3d. %s\n", i, t)
	}*/

	// Parse it
	err = t.parse()
	if err != nil {
		return nil, err
	}

	return t, nil
}

func (tpl *Template) Execute(context *Context) (string, error) {
	// Determine the parent to be executed (for template inheritance)
	parent := tpl
	for parent.parent != nil {
		parent = parent.parent
	}

	// Create context if none is given
	if context == nil {
		context = &Context{}
	} else {
		err := context.checkForValidIdentifiers()
		if err != nil {
			return "", err
		}
	}

	// Create operational context
	ctx := &ExecutionContext{
		template:    parent,
		Public:      context,
		Private:     &Context{},
		StringStore: make(map[string]string),
	}

	// Run the selected document
	return parent.root.Execute(ctx)
}

// Executes the template with the given context and writes to http.ResponseWriter
// on success. Context can be nil. Nothing is written on error; instead the error
// is being returned.
func (tpl *Template) ExecuteRW(w http.ResponseWriter, context *Context) error {
	s, err := tpl.Execute(context)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(s))
	if err != nil {
		return err
	}
	return nil
}
