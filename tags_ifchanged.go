package pongo2

import (
	"bytes"
)

// tagIfchangedNode represents the {% ifchanged %} tag.
//
// The ifchanged tag checks if a value has changed from the previous iteration
// in a loop. It's useful for displaying grouped data or section headers.
//
// Basic usage (checks if block content changed):
//
//	{% for date in days %}
//	    {% ifchanged %}{{ date.month }}{% endifchanged %}
//	    {{ date.day }}
//	{% endfor %}
//
// Watching specific variables:
//
//	{% for item in items %}
//	    {% ifchanged item.category %}
//	        <h2>{{ item.category }}</h2>
//	    {% endifchanged %}
//	    <p>{{ item.name }}</p>
//	{% endfor %}
//
// Using else clause (rendered when value hasn't changed):
//
//	{% for item in items %}
//	    {% ifchanged item.section %}
//	        <h3>{{ item.section }}</h3>
//	    {% else %}
//	        <hr>
//	    {% endifchanged %}
//	    {{ item.name }}
//	{% endfor %}
//
// Watching multiple variables:
//
//	{% for item in items %}
//	    {% ifchanged item.year item.month %}
//	        <h2>{{ item.year }}-{{ item.month }}</h2>
//	    {% endifchanged %}
//	{% endfor %}
type tagIfchangedNode struct {
	watchedExpr []IEvaluator
	lastValues  []*Value
	lastContent []byte
	thenWrapper *NodeWrapper
	elseWrapper *NodeWrapper
}

// Execute checks if watched expressions (or rendered content) have changed
// since the last call. Renders the then block if changed, else block otherwise.
func (node *tagIfchangedNode) Execute(ctx *ExecutionContext, writer TemplateWriter) error {
	if len(node.watchedExpr) == 0 {
		// Check against own rendered body

		// TODO: Check opportunity for buffer recycling
		buf := bytes.NewBuffer(make([]byte, 0, 1024)) // 1 KiB

		err := node.thenWrapper.Execute(ctx, buf)
		if err != nil {
			return err
		}

		bufBytes := buf.Bytes()
		if !bytes.Equal(node.lastContent, bufBytes) {
			// Rendered content changed, output it
			if _, err := writer.Write(bufBytes); err != nil {
				return err
			}
			node.lastContent = bufBytes
		} else if node.elseWrapper != nil {
			// Content hasn't changed, render else block if present
			if err := node.elseWrapper.Execute(ctx, writer); err != nil {
				return err
			}
		}
	} else {
		nowValues := make([]*Value, 0, len(node.watchedExpr))
		for _, expr := range node.watchedExpr {
			val, err := expr.Evaluate(ctx)
			if err != nil {
				return err
			}
			nowValues = append(nowValues, val)
		}

		// Compare old to new values now
		changed := len(node.lastValues) == 0

		for idx, oldVal := range node.lastValues {
			if !oldVal.EqualValueTo(nowValues[idx]) {
				changed = true
				break // we can stop here because ONE value changed
			}
		}

		node.lastValues = nowValues

		if changed {
			// Render thenWrapper
			err := node.thenWrapper.Execute(ctx, writer)
			if err != nil {
				return err
			}
		} else {
			// Render elseWrapper
			if node.elseWrapper != nil {
				err := node.elseWrapper.Execute(ctx, writer)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// tagIfchangedParser parses the {% ifchanged %} tag. It accepts zero or more
// expressions to watch; if none are given, it watches the rendered content.
func tagIfchangedParser(doc *Parser, start *Token, arguments *Parser) (INodeTag, error) {
	ifchangedNode := &tagIfchangedNode{}

	for arguments.Remaining() > 0 {
		// Parse condition
		expr, err := arguments.ParseExpression()
		if err != nil {
			return nil, err
		}
		ifchangedNode.watchedExpr = append(ifchangedNode.watchedExpr, expr)
	}

	if arguments.Remaining() > 0 {
		return nil, arguments.Error("Ifchanged-arguments are malformed.", nil)
	}

	// Wrap then/else-blocks
	wrapper, endargs, err := doc.WrapUntilTag("else", "endifchanged")
	if err != nil {
		return nil, err
	}
	ifchangedNode.thenWrapper = wrapper

	if endargs.Count() > 0 {
		return nil, endargs.Error("Arguments not allowed here.", nil)
	}

	if wrapper.Endtag == "else" {
		// if there's an else in the if-statement, we need the else-Block as well
		wrapper, endargs, err = doc.WrapUntilTag("endifchanged")
		if err != nil {
			return nil, err
		}
		ifchangedNode.elseWrapper = wrapper

		if endargs.Count() > 0 {
			return nil, endargs.Error("Arguments not allowed here.", nil)
		}
	}

	return ifchangedNode, nil
}

func init() {
	mustRegisterTag("ifchanged", tagIfchangedParser)
}
