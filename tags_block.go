package pongo2

import (
	"bytes"
	"fmt"
)

// tagBlockNode represents the {% block %} tag.
//
// The block tag defines a block that can be overridden by child templates.
// Blocks are used in conjunction with {% extends %} to implement template inheritance.
// Child templates can override blocks defined in parent templates.
//
// Usage in base template (base.html):
//
//	<html>
//	<head><title>{% block title %}Default Title{% endblock %}</title></head>
//	<body>
//	    {% block content %}Default content{% endblock %}
//	</body>
//	</html>
//
// Usage in child template:
//
//	{% extends "base.html" %}
//	{% block title %}My Page Title{% endblock %}
//	{% block content %}
//	    <h1>Welcome!</h1>
//	    <p>This is my custom content.</p>
//	{% endblock %}
//
// Using block.Super() to include parent content:
//
//	{% extends "base.html" %}
//	{% block content %}
//	    {{ block.Super }}
//	    <p>Additional content after parent's content.</p>
//	{% endblock %}
//
// The endblock tag can optionally include the block name for clarity:
//
//	{% block sidebar %}...{% endblock sidebar %}
type tagBlockNode struct {
	name string
}

// getBlockWrappers collects all block wrappers with the same name from the
// template inheritance chain. It walks from the current template down through
// all child templates, gathering overriding block definitions.
func (node *tagBlockNode) getBlockWrappers(tpl *Template) []*NodeWrapper {
	nodeWrappers := make([]*NodeWrapper, 0)
	var t *NodeWrapper

	for tpl != nil {
		t = tpl.blocks[node.name]
		if t != nil {
			nodeWrappers = append(nodeWrappers, t)
		}
		tpl = tpl.child
	}

	return nodeWrappers
}

// Execute renders the most-derived (child-most) version of this block.
// It sets up the block context with Super() support for accessing parent blocks.
func (node *tagBlockNode) Execute(ctx *ExecutionContext, writer TemplateWriter) error {
	tpl := ctx.template
	if tpl == nil {
		panic("internal error: tpl == nil")
	}

	// Determine the block to execute
	blockWrappers := node.getBlockWrappers(tpl)
	lenBlockWrappers := len(blockWrappers)

	if lenBlockWrappers == 0 {
		return ctx.Error("internal error: len(block_wrappers) == 0 in tagBlockNode.Execute()", nil)
	}

	blockWrapper := blockWrappers[lenBlockWrappers-1]
	ctx.Private["block"] = tagBlockInformation{
		ctx:      ctx,
		wrappers: blockWrappers[0 : lenBlockWrappers-1],
	}
	err := blockWrapper.Execute(ctx, writer)
	if err != nil {
		return err
	}

	return nil
}

// tagBlockInformation holds block context during execution, providing
// access to parent block content via the Super() method.
type tagBlockInformation struct {
	ctx      *ExecutionContext
	wrappers []*NodeWrapper
}

// Super renders and returns the parent block's content. This allows child
// templates to include the parent's block content within their override.
// Returns an empty safe value if there is no parent block.
func (t tagBlockInformation) Super() (*Value, error) {
	lenWrappers := len(t.wrappers)

	if lenWrappers == 0 {
		return AsSafeValue(""), nil
	}

	superCtx := NewChildExecutionContext(t.ctx)
	superCtx.Private["block"] = tagBlockInformation{
		ctx:      t.ctx,
		wrappers: t.wrappers[0 : lenWrappers-1],
	}

	blockWrapper := t.wrappers[lenWrappers-1]
	buf := bytes.NewBufferString("")
	err := blockWrapper.Execute(superCtx, &templateWriter{buf})
	if err != nil {
		return AsSafeValue(""), err
	}
	return AsSafeValue(buf.String()), nil
}

// tagBlockParser parses the {% block %} tag. It requires an identifier
// for the block name and registers the block in the template's block map.
func tagBlockParser(doc *Parser, start *Token, arguments *Parser) (INodeTag, error) {
	if arguments.Count() == 0 {
		return nil, arguments.Error("Tag 'block' requires an identifier.", nil)
	}

	nameToken := arguments.MatchType(TokenIdentifier)
	if nameToken == nil {
		return nil, arguments.Error("First argument for tag 'block' must be an identifier.", nil)
	}

	if arguments.Remaining() != 0 {
		return nil, arguments.Error("Tag 'block' takes exactly 1 argument (an identifier).", nil)
	}

	wrapper, endtagargs, err := doc.WrapUntilTag("endblock")
	if err != nil {
		return nil, err
	}
	if endtagargs.Remaining() > 0 {
		endtagnameToken := endtagargs.MatchType(TokenIdentifier)
		if endtagnameToken != nil {
			if endtagnameToken.Val != nameToken.Val {
				return nil, endtagargs.Error(fmt.Sprintf("Name for 'endblock' must equal to 'block'-tag's name ('%s' != '%s').",
					nameToken.Val, endtagnameToken.Val), nil)
			}
		}

		if endtagnameToken == nil || endtagargs.Remaining() > 0 {
			return nil, endtagargs.Error("Either no or only one argument (identifier) allowed for 'endblock'.", nil)
		}
	}

	tpl := doc.template
	if tpl == nil {
		panic("internal error: tpl == nil")
	}
	_, hasBlock := tpl.blocks[nameToken.Val]
	if !hasBlock {
		tpl.blocks[nameToken.Val] = wrapper
	} else {
		return nil, arguments.Error(fmt.Sprintf("Block named '%s' already defined", nameToken.Val), nil)
	}

	return &tagBlockNode{name: nameToken.Val}, nil
}

func init() {
	mustRegisterTag("block", tagBlockParser)
}
