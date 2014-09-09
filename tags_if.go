package pongo2

import (
	"bytes"
)

type tagIfNode struct {
	conditions	[]IEvaluator
	wrappers	[]*NodeWrapper
}

func (node *tagIfNode) Execute(ctx *ExecutionContext, buffer *bytes.Buffer) error {
	for i, condition := range node.conditions {
		result, err := condition.Evaluate(ctx)
		if err != nil {
			return err
		}

		if result.IsTrue() {
			return node.wrappers[i].Execute(ctx, buffer)
		} else {
			// Last condition?
			if len(node.conditions) == i + 1 && len(node.wrappers) > i + 1 {
				return node.wrappers[i + 1].Execute(ctx, buffer)
			}
		}
	}
	return nil
}

func tagIfParser(doc *Parser, start *Token, arguments *Parser) (INodeTag, error) {
	if_node := &tagIfNode{}

	// Parse first and main IF condition
	condition, err := arguments.ParseExpression()
	if err != nil {
		return nil, err
	}
	if_node.conditions = append(if_node.conditions, condition)

	if arguments.Remaining() > 0 {
		return nil, arguments.Error("If-condition is malformed.", nil)
	}

	// Check the rest
	for {
		wrapper, args, err := doc.WrapUntilTag("else", "elseif", "endif")
		if err != nil {
			return nil, err
		}
		if_node.wrappers = append(if_node.wrappers, wrapper)

		if wrapper.Endtag == "elseif" {
			// ELSEIF can has condition
			condition, err := args.ParseExpression()
			if err != nil {
				return nil, err
			}
			if_node.conditions = append(if_node.conditions, condition)

			if args.Remaining() > 0 {
				return nil, args.Error("Elseif-condition is malformed.", nil)
			}
		} else if args.Count() > 0 {
			// ELSE and ENDIF can not
			return nil, args.Error("Arguments not allowed here.", nil)
		}

		if wrapper.Endtag == "endif" {
			break
		}
	}

	return if_node, nil
}

func init() {
	RegisterTag("if", tagIfParser)
}
