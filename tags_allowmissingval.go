package pongo2

import (
	"bytes"
)

type tagAllowMissingVal struct {
	position    *Token
	bodyWrapper *NodeWrapper
}

func (node *tagAllowMissingVal) Execute(ctx *ExecutionContext, writer TemplateWriter) *Error {
	temp := bytes.NewBuffer(make([]byte, 0, 1024)) // 1 KiB size
	ctx.AllowMissingVal = true

	err := node.bodyWrapper.Execute(ctx, temp)
	if err != nil {
		return err
	}
	templateSet := ctx.template.set
	currentTemplate, err2 := templateSet.FromBytes(temp.Bytes())
	if err2 != nil {
		return err2.(*Error)
	}
	if err := currentTemplate.root.Execute(ctx, writer); err != nil {
		return err
	}
	return nil
}

func tagHandleParser(doc *Parser, start *Token, arguments *Parser) (INodeTag, *Error) {
	execNode := &tagAllowMissingVal{
		position: start,
	}

	wrapper, _, err := doc.WrapUntilTag("endallowmissingval")
	if err != nil {
		return nil, err
	}
	execNode.bodyWrapper = wrapper

	return execNode, nil
}

func init() {
	MustRegisterTag("allowmissingval", tagHandleParser)
}
