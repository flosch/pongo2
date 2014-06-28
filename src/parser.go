package pongo2

import (
	"errors"
	"fmt"
	//"strings"
)

type IEvaluator interface {
	Evaluate(*ExecutionContext) (*Value, error)
}

type INode interface {
	Execute(*ExecutionContext) (string, error)
}

type INodeEvaluator interface {
	INode
	IEvaluator
}

type Parser struct {
	name   string
	idx    int
	tokens []*Token
}

// Creates a new parser to parse tokens.
// Used inside pongo2 to parse documents and to provide an easy-to-use
// parser for tag authors
func newParser(name string, tokens []*Token) *Parser {
	return &Parser{
		name:   name,
		tokens: tokens,
	}
}

func (p *Parser) Consume() {
	p.ConsumeN(1)
}

func (p *Parser) ConsumeN(count int) {
	p.idx += count
}

func (p *Parser) Current() *Token {
	return p.Get(p.idx)
}

func (p *Parser) MatchType(typ TokenType) *Token {
	if t := p.PeekType(typ); t != nil {
		p.Consume()
		return t
	}
	return nil
}

func (p *Parser) PeekType(typ TokenType) *Token {
	return p.PeekTypeN(0, typ)
}

func (p *Parser) PeekTypeN(shift int, typ TokenType) *Token {
	t := p.Get(p.idx + shift)
	if t != nil {
		if t.Typ == typ {
			return t
		}
	}
	return nil
}

func (p *Parser) Match(typ TokenType, val string) *Token {
	if t := p.Peek(typ, val); t != nil {
		p.Consume()
		return t
	}
	return nil
}

func (p *Parser) MatchOne(typ TokenType, vals ...string) *Token {
	for _, val := range vals {
		if t := p.Peek(typ, val); t != nil {
			p.Consume()
			return t
		}
	}
	return nil
}

func (p *Parser) Peek(typ TokenType, val string) *Token {
	return p.PeekN(0, typ, val)
}

func (p *Parser) PeekOne(typ TokenType, vals ...string) *Token {
	for _, v := range vals {
		t := p.PeekN(0, typ, v)
		if t != nil {
			return t
		}
	}
	return nil
}

func (p *Parser) PeekN(shift int, typ TokenType, val string) *Token {
	t := p.Get(p.idx + shift)
	if t != nil {
		if t.Typ == typ && t.Val == val {
			return t
		}
	}
	return nil
}

func (p *Parser) Remaining() int {
	return len(p.tokens) - p.idx
}

func (p *Parser) Count() int {
	return len(p.tokens)
}

func (p *Parser) Get(i int) *Token {
	if i < len(p.tokens) {
		return p.tokens[i]
	}
	return nil
}

func (p *Parser) GetR(shift int) *Token {
	i := p.idx + shift
	return p.Get(i)
}

func (p *Parser) Error(msg string, token *Token) error {
	if token == nil {
		// Set current token
		token = p.Current()
		if token == nil {
			// Set to last token
			if len(p.tokens) > 0 {
				token = p.tokens[len(p.tokens)-1]
			}
		}
	}
	pos := ""
	if token != nil {
		// No tokens available
		// TODO: Add location (from where?)
		pos = fmt.Sprintf(" | Line %d Col %d (%s)",
			token.Line, token.Col, token.String())
	}
	return errors.New(
		fmt.Sprintf("[Parse Error in %s%s] %s",
			p.name, pos, msg,
		))
}

// Wraps all nodes between starting tag and "{% endtag %}" and provides
// one simple interface to execute the wrapped nodes
func (p *Parser) WrapUntilTag(names ...string) (*NodeWrapper, error) {
	wrapper := &NodeWrapper{}

wrappingLoop:
	for p.Remaining() > 0 {
		// New tag, check whether we have to stop wrapping here
		if p.Peek(TokenSymbol, "{%") != nil {
			tag_ident := p.PeekTypeN(1, TokenIdentifier)

			if tag_ident != nil {
				// We've found a (!) end-tag

				found := false
				name := ""
				for _, n := range names {
					if tag_ident.Val == n {
						name = n
						found = true
						break
					}
				}

				// We only process the tag if we've found an end tag
				if found {
					if p.PeekN(2, TokenSymbol, "%}") != nil {
						// Okay, end the wrapping here
						wrapper.Endtag = tag_ident.Val

						p.ConsumeN(3)
						break wrappingLoop
					} else {
						// Arguments provided, which is not allowed
						return nil, p.Error(fmt.Sprintf("No arguments allowed for tag '%s'", name), tag_ident)
					}
				}
				/* else {
					// unexpected EOF
					return nil, p.Error(fmt.Sprintf("Unexpected EOF, tag '%s' not closed", name), tpl.tokens[tpl.parser_idx+1])
				}*/
			}

		}
		/*else {
			// unexpected EOF
			return nil, p.Error(fmt.Sprintf("Unexpected EOF (expected end-tags '%s')", strings.Join(names, ", ")), t)
		}*/

		// Otherwise process next element to be wrapped
		node, err := p.parseDocElement()
		if err != nil {
			return nil, err
		}
		wrapper.nodes = append(wrapper.nodes, node)
	}

	return wrapper, nil
}
