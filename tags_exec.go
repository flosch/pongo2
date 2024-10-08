package pongo2

import (
	"bytes"
	"fmt"
)

type tagExecNode struct {
	position    *Token
	bodyWrapper *NodeWrapper
}

func (node *tagExecNode) Execute(ctx *ExecutionContext, writer TemplateWriter) *Error {
	temp := bytes.NewBuffer(make([]byte, 0, 1024)) // 1 KiB size

	err := node.bodyWrapper.Execute(ctx, temp)
	if err != nil {
		return err
	}
	templateSet := ctx.template.set
	currentTemplate, err2 := templateSet.FromBytes(temp.Bytes())
	if err2 != nil {
		return err2.(*Error)
	}

	err = currentTemplate.root.Execute(ctx, writer)
	if err != nil {
		err.OrigError = fmt.Errorf("running exec tag: %w", err.OrigError)
		return err
	}

	return nil
}

func tagExecuteParser(doc *Parser, start *Token, arguments *Parser) (INodeTag, *Error) {
	execNode := &tagExecNode{
		position: start,
	}

	wrapper, _, err := doc.WrapUntilTag("endexec")
	if err != nil {
		return nil, err
	}
	execNode.bodyWrapper = wrapper

	return execNode, nil
}

func init() {
	MustRegisterTag("exec", tagExecuteParser)
}
