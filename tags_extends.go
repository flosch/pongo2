package pongo2

import (
	"fmt"
	"path/filepath"
)

type tagExtendsNode struct {
	filename_evaluator IEvaluator
	filename           string
	lazy               bool
}

func (node *tagExtendsNode) Execute(ctx *ExecutionContext) (string, error) {
	return "", nil
}

func tagExtendsParser(doc *Parser, start *Token, arguments *Parser) (INodeTag, error) {
	extends_node := &tagExtendsNode{}

	if doc.template.level > 1 {
		return nil, arguments.Error("The 'extends' tag can only defined on root level.", start)
	}

	if doc.template.parent != nil {
		// Already one parent
		return nil, arguments.Error("This template has already one parent.", start)
	}

	if arguments.PeekType(TokenString) == nil {
		// We will do lazy evaluation of the filename because no string is given
		filename_evaluator, err := arguments.ParseExpression()
		if err != nil {
			return nil, err
		}
		extends_node.filename_evaluator = filename_evaluator
		extends_node.lazy = true
	} else {
		// prepared, static template (because string is given)
		filename_str := arguments.MatchType(TokenString)

		// Get parent's filename relative to the child's template directory
		childs_dir := filepath.Dir(doc.template.name)
		parent_filename := filepath.Join(childs_dir, filename_str.Val)

		// Parse the parent
		parent_template, err := FromFile(parent_filename)
		if err != nil {
			return nil, err
		}

		// Keep track of things
		parent_template.child = doc.template
		doc.template.parent = parent_template
		extends_node.filename = parent_filename
	}

	if arguments.Remaining() > 0 {
		fmt.Printf("%+v\n", arguments.tokens)
		return nil, arguments.Error("Tag 'extends' does only take 1 argument.", nil)
	}

	return extends_node, nil
}

func init() {
	RegisterTag("extends", tagExtendsParser)
}
