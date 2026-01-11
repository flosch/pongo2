package pongo2

import (
	"fmt"
	"math"
)

// tagWidthratioNode represents the {% widthratio %} tag.
//
// The widthratio tag calculates a ratio and multiplies it by a constant,
// useful for creating bar charts, progress indicators, and other proportional
// visualizations.
//
// Syntax: {% widthratio current_value max_value max_width %}
//
// The formula is: ceil(current_value / max_value * max_width)
//
// Basic usage (progress bar):
//
//	<div class="progress-bar" style="width: {% widthratio task.completed task.total 100 %}%">
//	</div>
//
// If task.completed=75 and task.total=100, output: "width: 75%"
//
// Creating a bar chart:
//
//	{% for item in items %}
//	    <div style="width: {% widthratio item.value max_value 200 %}px">
//	        {{ item.name }}
//	    </div>
//	{% endfor %}
//
// Storing result in a variable using "as":
//
//	{% widthratio current max 100 as percentage %}
//	<p>Progress: {{ percentage }}%</p>
//
// Example calculations:
//
//	{% widthratio 50 100 200 %}   {# Output: 100 (50/100 * 200) #}
//	{% widthratio 75 100 400 %}   {# Output: 300 (75/100 * 400) #}
//	{% widthratio 1 3 100 %}      {# Output: 34 (1/3 * 100, rounded up) #}
type tagWidthratioNode struct {
	position     *Token
	current, max IEvaluator
	width        IEvaluator
	ctxName      string
}

func (node *tagWidthratioNode) Execute(ctx *ExecutionContext, writer TemplateWriter) error {
	current, err := node.current.Evaluate(ctx)
	if err != nil {
		return err
	}

	max, err := node.max.Evaluate(ctx)
	if err != nil {
		return err
	}

	width, err := node.width.Evaluate(ctx)
	if err != nil {
		return err
	}

	value := int(math.Ceil(current.Float()/max.Float()*width.Float() + 0.5))

	if node.ctxName == "" {
		writer.WriteString(fmt.Sprintf("%d", value))
	} else {
		ctx.Private[node.ctxName] = value
	}

	return nil
}

func tagWidthratioParser(doc *Parser, start *Token, arguments *Parser) (INodeTag, error) {
	widthratioNode := &tagWidthratioNode{
		position: start,
	}

	current, err := arguments.ParseExpression()
	if err != nil {
		return nil, err
	}
	widthratioNode.current = current

	max, err := arguments.ParseExpression()
	if err != nil {
		return nil, err
	}
	widthratioNode.max = max

	width, err := arguments.ParseExpression()
	if err != nil {
		return nil, err
	}
	widthratioNode.width = width

	if arguments.MatchOne(TokenKeyword, "as") != nil {
		// Name follows
		nameToken := arguments.MatchType(TokenIdentifier)
		if nameToken == nil {
			return nil, arguments.Error("Expected name (identifier).", nil)
		}
		widthratioNode.ctxName = nameToken.Val
	}

	if arguments.Remaining() > 0 {
		return nil, arguments.Error("Malformed widthratio-tag arguments.", nil)
	}

	return widthratioNode, nil
}

func init() {
	RegisterTag("widthratio", tagWidthratioParser)
}
