package pongo2

import (
	"bytes"
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
	newContext := make(Context)
	newContext.Update(ctx.Private)
	newContext.Update(ctx.Public)
	s := temp.String()
	currentTemplate, _ := templateSet.FromString(s)
	finalRes, _ := currentTemplate.Execute(newContext)
	moveMacrosToMainTemplate(currentTemplate, ctx)

	_, err2 := writer.WriteString(finalRes)
	if err2 != nil {
		return nil
	}

	return nil
}

func moveMacrosToMainTemplate(template *Template, context *ExecutionContext) {
	for _, macro := range template.exportedMacros {
		macro.Execute(context, nil)
	}
	for _, node := range template.root.Nodes {
		importNode, ok := node.(*tagImportNode)
		if ok {
			for _, macro := range importNode.macros {
				macro.Execute(context, nil)
			}
		}
		macroNode, ok := node.(*tagMacroNode)
		if ok {
			macroNode.Execute(context, nil)
		}
	}
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
	RegisterTag("exec", tagExecuteParser)
}
