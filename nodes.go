package pongo2

import (
	"strings"
)

// The root document
type nodeDocument struct {
	Nodes []INode
}

func (doc *nodeDocument) Execute(ctx *ExecutionContext) (string, error) {
	collection := make([]string, 0, len(doc.Nodes))
	for _, n := range doc.Nodes {
		buf, err := n.Execute(ctx)
		if err != nil {
			return "", err
		}
		collection = append(collection, buf)
	}
	return strings.Join(collection, ""), nil
}
