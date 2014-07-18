package pongo2

import (
	"bytes"
)

type nodeHTML struct {
	token *Token
}

func (n *nodeHTML) Execute(ctx *ExecutionContext, buffer *bytes.Buffer) error {
	buffer.WriteString(n.token.Val)
	return nil
}
