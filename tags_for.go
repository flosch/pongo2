package pongo2

import (
	"strings"
)

type tagForNode struct {
	key              string
	value            string // only for maps: for key, value in map
	object_evaluator INodeEvaluator

	bodyWrapper  *NodeWrapper
	emptyWrapper *NodeWrapper
}

type tagForLoopInformation struct {
	Counter     int
	Counter0    int
	Revcounter  int
	Revcounter0 int
	First       bool
	Last        bool
	Parentloop  *tagForLoopInformation
}

func (node *tagForNode) Execute(ctx *ExecutionContext) (s string, forError error) {
	// Backup forloop (as parentloop in public context), key-name and value-name
	parentloop := ctx.Private["forloop"]
	backup_key := ctx.Private[node.key]
	backup_value := ctx.Private[node.value]

	// Create loop struct
	loopInfo := &tagForLoopInformation{
		First: true,
	}

	// Is it a loop in a loop?
	if parentloop != nil {
		loopInfo.Parentloop = parentloop.(*tagForLoopInformation)
	}

	// Register loopInfo in public context
	ctx.Private["forloop"] = loopInfo

	container := make([]string, 0)

	obj, err := node.object_evaluator.Evaluate(ctx)
	if err != nil {
		return "", err
	}

	obj.Iterate(func(idx, count int, key, value *Value) bool {
		// There's something to iterate over (correct type and at least 1 item)

		// Update loop infos and public context
		ctx.Private[node.key] = key
		if value != nil {
			ctx.Private[node.value] = value
		}
		loopInfo.Counter = idx + 1
		loopInfo.Counter0 = idx
		if idx == 1 {
			loopInfo.First = false
		}
		if idx+1 == count {
			loopInfo.Last = true
		}
		loopInfo.Revcounter = count - idx        // TODO: Not sure about this, have to look it up
		loopInfo.Revcounter0 = count - (idx + 1) // TODO: Not sure about this, have to look it up

		// Render elements with updated context
		s, err := node.bodyWrapper.Execute(ctx)
		if err != nil {
			forError = err
			return false
		}
		container = append(container, s)
		return true
	}, func() {
		// Nothing to iterate over (maybe wrong type or no items)
		if node.emptyWrapper != nil {
			s, err := node.emptyWrapper.Execute(ctx)
			if err != nil {
				forError = err
			}
			container = append(container, s)
		}
	})

	if forError != nil {
		return
	}

	// Restore forloop and parentloop
	ctx.Private[node.key] = backup_key
	if backup_value != nil {
		ctx.Private[node.value] = backup_value
	}
	ctx.Private["forloop"] = parentloop

	// Return the rendered template
	return strings.Join(container, ""), nil
}

func tagForParser(doc *Parser, start *Token, arguments *Parser) (INodeTag, error) {
	for_node := &tagForNode{}

	// Arguments parsing
	var value_token *Token
	key_token := arguments.MatchType(TokenIdentifier)
	if key_token == nil {
		return nil, arguments.Error("Expected an key identifier as first argument for 'for'-tag", nil)
	}

	if arguments.Match(TokenSymbol, ",") != nil {
		// Value name is provided
		value_token = arguments.MatchType(TokenIdentifier)
		if value_token == nil {
			return nil, arguments.Error("Value name must be an identifier.", nil)
		}
	}

	if arguments.Match(TokenKeyword, "in") == nil {
		return nil, arguments.Error("Expected keyword 'in'.", nil)
	}

	object_evaluator, err := arguments.ParseExpression()
	if err != nil {
		return nil, err
	}
	for_node.object_evaluator = object_evaluator
	for_node.key = key_token.Val
	if value_token != nil {
		for_node.value = value_token.Val
	}

	if arguments.Remaining() > 0 {
		return nil, arguments.Error("Malformed for-loop arguments.", nil)
	}

	// Body wrapping
	wrapper, err := doc.WrapUntilTag("empty", "endfor")
	if err != nil {
		return nil, err
	}
	for_node.bodyWrapper = wrapper

	if wrapper.Endtag == "empty" {
		// if there's an else in the if-statement, we need the else-Block as well
		wrapper, err = doc.WrapUntilTag("endfor")
		if err != nil {
			return nil, err
		}

		for_node.emptyWrapper = wrapper
	}

	return for_node, nil
}

func init() {
	RegisterTag("for", tagForParser)
}
