package pongo2

type NodeHTML struct {
	token *Token
}

func (n *NodeHTML) Execute(ctx *ExecutionContext) (string, error) {
	return n.token.Val, nil
}
