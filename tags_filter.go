package pongo2

import (
	"bytes"
)

// nodeFilterCall represents a single filter call with its name and optional parameter.
type nodeFilterCall struct {
	name      string
	paramExpr IEvaluator
}

// tagFilterNode represents the {% filter %} tag.
//
// The filter tag applies one or more filters to a block of template content.
// This is useful when you want to apply a filter to a large block of text
// rather than a single variable.
//
// Usage with a single filter:
//
//	{% filter upper %}
//	    This text will be converted to uppercase.
//	{% endfilter %}
//
// Output: "THIS TEXT WILL BE CONVERTED TO UPPERCASE."
//
// Usage with filter parameters:
//
//	{% filter truncatewords:3 %}
//	    This is a longer text that will be truncated.
//	{% endfilter %}
//
// Output: "This is a ..."
//
// Chaining multiple filters:
//
//	{% filter lower|capfirst %}
//	    THIS TEXT WILL BE LOWERCASED THEN CAPITALIZED.
//	{% endfilter %}
//
// Output: "This text will be lowercased then capitalized."
//
// Combining escape and linebreaksbr:
//
//	{% filter escape|linebreaksbr %}
//	Line 1
//	Line 2
//	{% endfilter %}
//
// Output: "Line 1<br />Line 2"
type tagFilterNode struct {
	position    *Token
	bodyWrapper *NodeWrapper
	filterChain []*nodeFilterCall
}

// Execute renders the block content, then applies the filter chain to the
// result. Each filter transforms the output of the previous one.
func (node *tagFilterNode) Execute(ctx *ExecutionContext, writer TemplateWriter) error {
	temp := bytes.NewBuffer(make([]byte, 0, 1024)) // 1 KiB size

	err := node.bodyWrapper.Execute(ctx, temp)
	if err != nil {
		return err
	}

	value := AsValue(temp.String())

	for _, call := range node.filterChain {
		var param *Value
		if call.paramExpr != nil {
			param, err = call.paramExpr.Evaluate(ctx)
			if err != nil {
				return err
			}
		} else {
			param = AsValue(nil)
		}
		value, err = ctx.template.set.ApplyFilter(call.name, value, param)
		if err != nil {
			return ctx.Error(err.Error(), node.position)
		}
	}

	_, err = writer.WriteString(value.String())
	return err
}

// tagFilterParser parses the {% filter %} tag. It requires at least one filter
// name and supports filter chaining with | and parameters with :.
func tagFilterParser(doc *Parser, start *Token, arguments *Parser) (INodeTag, error) {
	filterNode := &tagFilterNode{
		position: start,
	}

	wrapper, _, err := doc.WrapUntilTag("endfilter")
	if err != nil {
		return nil, err
	}
	filterNode.bodyWrapper = wrapper

	// Django requires at least one filter
	if arguments.Count() == 0 {
		return nil, arguments.Error("Tag 'filter' requires at least one filter.", nil)
	}

	for arguments.Remaining() > 0 {
		filterCall := &nodeFilterCall{}

		nameToken := arguments.MatchType(TokenIdentifier)
		if nameToken == nil {
			return nil, arguments.Error("Expected a filter name (identifier).", nil)
		}
		filterCall.name = nameToken.Val

		if arguments.MatchOne(TokenSymbol, ":") != nil {
			// Filter parameter
			// NOTICE: we can't use ParseExpression() here, because it would parse the next filter "|..." as well in the argument list
			expr, err := arguments.parseVariableOrLiteral()
			if err != nil {
				return nil, err
			}
			filterCall.paramExpr = expr
		}

		filterNode.filterChain = append(filterNode.filterChain, filterCall)

		if arguments.MatchOne(TokenSymbol, "|") == nil {
			break
		}
	}

	if arguments.Remaining() > 0 {
		return nil, arguments.Error("Malformed filter-tag arguments.", nil)
	}

	return filterNode, nil
}

func init() {
	mustRegisterTag("filter", tagFilterParser)
}
