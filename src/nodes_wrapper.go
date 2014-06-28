package pongo2

import (
	"strings"
)

type NodeWrapper struct {
	Endtag string
	nodes  []INode
}

func (wrapper *NodeWrapper) Execute(ctx *ExecutionContext) (string, error) {
	container := make([]string, 0, len(wrapper.nodes))
	for _, n := range wrapper.nodes {
		buf, err := n.Execute(ctx)
		if err != nil {
			return "", err
		}
		container = append(container, buf)
	}
	return strings.Join(container, ""), nil
}
