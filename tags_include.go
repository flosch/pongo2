package pongo2

import (
	"path/filepath"
)

type tagIncludeNode struct {
	tpl      *Template
	filename string
}

func (node *tagIncludeNode) Execute(ctx *ExecutionContext) (string, error) {
	return node.tpl.Execute(ctx.Public)
}

func tagIncludeParser(doc *Parser, start *Token, arguments *Parser) (INodeTag, error) {
	include_node := &tagIncludeNode{}

	if filename_token := arguments.MatchType(TokenString); filename_token != nil {
		// prepared, static template

		// Get parent's filename relative to the child's template directory
		childs_dir := filepath.Dir(doc.template.name)
		parent_filename := filepath.Join(childs_dir, filename_token.Val)

		// Parse the parent
		include_node.filename = parent_filename
		included_tpl, err := FromFile(parent_filename)
		if err != nil {
			return nil, err
		}
		include_node.tpl = included_tpl
	} else {
		return nil, arguments.Error("Tag 'extends' requires a template filename as string.", nil)
	}

	if arguments.Remaining() > 0 {
		return nil, arguments.Error("Tag 'include' does only take 1 argument.", nil)
	}

	return include_node, nil
}

func init() {
	RegisterTag("include", tagIncludeParser)
}
