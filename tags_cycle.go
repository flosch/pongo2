package pongo2

import "fmt"

type tagCycleValue struct {
	node  *tagCycleNode
	value *Value
}

type tagCycleNode struct {
	position *Token
	args     []IEvaluator
	idx      int
	asName   string
	silent   bool
}

func (cv *tagCycleValue) String() string {
	return cv.value.String()
}

func (node *tagCycleNode) Execute(ctx *ExecutionContext, writer TemplateWriter) *Error {
	item := node.args[node.idx%len(node.args)]
	node.idx++

	val, err := item.Evaluate(ctx)
	if err != nil {
		return err
	}

	if t, ok := val.Interface().(*tagCycleValue); ok {
		// {% cycle "test1" "test2"
		// {% cycle cycleitem %}

		// Update the cycle value with next value
		item := t.node.args[t.node.idx%len(t.node.args)]
		t.node.idx++

		val, err := item.Evaluate(ctx)
		if err != nil {
			return err
		}

		t.value = val

		if !t.node.silent {
			if _, err := writer.WriteString(val.String()); err != nil {
				return ctx.Error(err, node.position)
			}
		}
	} else {
		// Regular call

		cycleValue := &tagCycleValue{
			node:  node,
			value: val,
		}

		if node.asName != "" {
			ctx.Private[node.asName] = cycleValue
		}
		if !node.silent {
			if _, err := writer.WriteString(val.String()); err != nil {
				return ctx.Error(err, node.position)
			}
		}
	}

	return nil
}

// HINT: We're not supporting the old comma-separated list of expressions argument-style
func tagCycleParser(doc *Parser, start *Token, arguments *Parser) (INodeTag, *Error) {
	cycleNode := &tagCycleNode{
		position: start,
	}

	for arguments.Remaining() > 0 {
		node, err := arguments.ParseExpression()
		if err != nil {
			return nil, err
		}
		cycleNode.args = append(cycleNode.args, node)

		if arguments.MatchOne(TokenKeyword, "as") != nil {
			// as

			nameToken := arguments.MatchType(TokenIdentifier)
			if nameToken == nil {
				return nil, arguments.Error(fmt.Errorf("Name (identifier) expected after 'as'."), nil)
			}
			cycleNode.asName = nameToken.Val

			if arguments.MatchOne(TokenIdentifier, "silent") != nil {
				cycleNode.silent = true
			}

			// Now we're finished
			break
		}
	}

	if arguments.Remaining() > 0 {
		return nil, arguments.Error(fmt.Errorf("Malformed cycle-tag."), nil)
	}

	return cycleNode, nil
}

func init() {
	MustRegisterTag("cycle", tagCycleParser)
}
