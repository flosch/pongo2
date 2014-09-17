package pongo2

import (
	"bytes"
)

type nodeFilterCall struct {
	name       string
	params []IEvaluator
}

type tagFilterNode struct {
	position    *Token
	bodyWrapper *NodeWrapper
	filterChain []*nodeFilterCall
}

func (node *tagFilterNode) Execute(ctx *ExecutionContext, buffer *bytes.Buffer) error {
	temp := bytes.NewBuffer(make([]byte, 0, 1024)) // 1 KiB size

	err := node.bodyWrapper.Execute(ctx, temp)
	if err != nil {
		return err
	}

	value := AsValue(temp.String())

	for _, call := range node.filterChain {
		var params []*Value
		if len(call.params) > 0 {
			for _, paramEvo := range call.params {
				param, err := paramEvo.Evaluate(ctx)
				if err != nil {
					return err
				}
				params = append(params, param)
			}
		}
		value, err = ApplyFilter(call.name, value, params...)
		if err != nil {
			return ctx.Error(err.Error(), node.position)
		}
	}

	buffer.WriteString(value.String())

	return nil
}

func tagFilterParser(doc *Parser, start *Token, arguments *Parser) (INodeTag, error) {
	filter_node := &tagFilterNode{
		position: start,
	}

	wrapper, _, err := doc.WrapUntilTag("endfilter")
	if err != nil {
		return nil, err
	}
	filter_node.bodyWrapper = wrapper

	for arguments.Remaining() > 0 {
		filterCall := &nodeFilterCall{}

		name_token := arguments.MatchType(TokenIdentifier)
		if name_token == nil {
			return nil, arguments.Error("Expected a filter name (identifier).", nil)
		}
		filterCall.name = name_token.Val

		for {
			if arguments.MatchOne(TokenSymbol, ":") == nil {
				break
			}
			// Filter parameter
			// NOTICE: we can't use ParseExpression() here, because it would parse the next filter "|..." as well in the argument list
			expr, err := arguments.parseVariableOrLiteral()
			if err != nil {
				return nil, err
			}
			filterCall.params = append(filterCall.params, expr)
		}

		filter_node.filterChain = append(filter_node.filterChain, filterCall)

		if arguments.MatchOne(TokenSymbol, "|") == nil {
			break
		}
	}

	if arguments.Remaining() > 0 {
		return nil, arguments.Error("Malformed filter-tag arguments.", nil)
	}

	return filter_node, nil
}

func init() {
	RegisterTag("filter", tagFilterParser)
}
