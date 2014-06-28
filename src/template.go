package pongo2

//import "fmt"

type Template struct {
	// Input
	name string
	tpl  string

	// Calculation
	tokens []*Token
	parser *Parser

	// Output
	root *NodeDocument
}

func newTemplateString(tpl string) (*Template, error) {
	return newTemplate("<string>", tpl)
}

func newTemplate(name, tpl string) (*Template, error) {
	// Create the template
	t := &Template{
		name: name,
		tpl:  tpl,
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
		template:    tpl,
		Public:      context,
		Private:     &Context{},
		StringStore: make(map[string]string),
	}

	// Run the document
	return tpl.root.Execute(ctx)
}
