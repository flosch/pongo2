package pongo2

/*
Missing built-in tags:
    autoescape
    csrf_token
    cycle
    debug
    filter
    firstof
    Boolean operators
    not in operator
    ifchanged
    load
    now
    regroup
    Grouping on other properties
    spaceless
    verbatim
    widthratio
    with

Following built-in tags wont be added:
	url
*/

import (
	"fmt"
)

type INodeTag interface {
	INode
}

type tagParser func(doc *Parser, start *Token, arguments *Parser) (INodeTag, error)

type tag struct {
	name   string
	parser tagParser
}

var tags map[string]*tag

func init() {
	tags = make(map[string]*tag)
}

func RegisterTag(name string, parserFn tagParser) {
	_, existing := tags[name]
	if existing {
		panic(fmt.Sprintf("Tag with name '%s' is already registered.", name))
	}
	tags[name] = &tag{
		name:   name,
		parser: parserFn,
	}
}

// Tag = "{%" IDENT ARGS "%}"
func (p *Parser) parseTagElement() (INodeTag, error) {
	p.Consume() // consume "{%"
	token_name := p.MatchType(TokenIdentifier)

	// Check for identifier
	if token_name == nil {
		return nil, p.Error("Tag name must be an identifier.", nil)
	}

	// Check for the existing tag
	tag, exists := tags[token_name.Val]
	if !exists {
		// Does not exists
		return nil, p.Error(fmt.Sprintf("Tag '%s' not found (or beginning tag not provided)", token_name.Val), token_name)
	}

	args_token := make([]*Token, 0)
	for p.Peek(TokenSymbol, "%}") == nil {
		// Add token to args
		args_token = append(args_token, p.Current())
		p.Consume() // next token
	}

	// EOF?
	if p.Remaining() == 0 {
		return nil, p.Error("Unexpectedly reached EOF, no tag end found.", nil)
	}

	p.Match(TokenSymbol, "%}")

	p.template.level++
	defer func() { p.template.level-- }()
	return tag.parser(p, token_name, newParser(p.name, args_token, nil))
}
