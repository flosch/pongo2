package pongo2

type nodeHTML struct {
	token *Token
}

func (n *nodeHTML) Execute(ctx *ExecutionContext) (string, error) {
	return n.token.Val, nil
}
