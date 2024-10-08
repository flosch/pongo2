package pongo2

import (
	"fmt"
	"time"
)

type tagNowNode struct {
	position *Token
	format   string
	fake     bool
}

func (node *tagNowNode) Execute(ctx *ExecutionContext, writer TemplateWriter) *Error {
	var t time.Time
	if node.fake {
		t = time.Date(2014, time.February, 0o5, 18, 31, 45, 0o0, time.UTC)
	} else {
		t = time.Now()
	}

	if _, err := writer.WriteString(t.Format(node.format)); err != nil {
		return ctx.Error(err, node.position)
	}

	return nil
}

func tagNowParser(doc *Parser, start *Token, arguments *Parser) (INodeTag, *Error) {
	nowNode := &tagNowNode{
		position: start,
	}

	formatToken := arguments.MatchType(TokenString)
	if formatToken == nil {
		return nil, arguments.Error(fmt.Errorf("Expected a format string."), nil)
	}
	nowNode.format = formatToken.Val

	if arguments.MatchOne(TokenIdentifier, "fake") != nil {
		nowNode.fake = true
	}

	if arguments.Remaining() > 0 {
		return nil, arguments.Error(fmt.Errorf("Malformed now-tag arguments."), nil)
	}

	return nowNode, nil
}

func init() {
	MustRegisterTag("now", tagNowParser)
}
