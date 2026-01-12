package pongo2

import (
	"bytes"
	"regexp"
)

// tagSpacelessNode represents the {% spaceless %} tag.
//
// The spaceless tag removes whitespace between HTML tags. It only removes
// whitespace that appears between closing and opening tags, not whitespace
// within tags or within text content.
//
// Basic usage:
//
//	{% spaceless %}
//	    <ul>
//	        <li>Item 1</li>
//	        <li>Item 2</li>
//	    </ul>
//	{% endspaceless %}
//
// Output: "<ul><li>Item 1</li><li>Item 2</li></ul>"
//
// Note that whitespace inside text content is preserved:
//
//	{% spaceless %}
//	    <p>
//	        Hello    World
//	    </p>
//	{% endspaceless %}
//
// Output: "<p>Hello    World</p>"
//
// Use cases:
//   - Minimizing HTML output size
//   - Removing unwanted whitespace in inline elements
//   - Cleaning up template-generated HTML
//
// Note: Only whitespace between tags is removed. Whitespace within tag
// content or attributes is preserved.
type tagSpacelessNode struct {
	wrapper *NodeWrapper
}

var tagSpacelessRegexp = regexp.MustCompile(`(?U:(<.*>))([\t\n\v\f\r ]+)(?U:(<.*>))`)

func (node *tagSpacelessNode) Execute(ctx *ExecutionContext, writer TemplateWriter) error {
	b := bytes.NewBuffer(make([]byte, 0, 1024)) // 1 KiB

	err := node.wrapper.Execute(ctx, b)
	if err != nil {
		return err
	}

	s := b.String()
	// Repeat this recursively
	changed := true
	for changed {
		s2 := tagSpacelessRegexp.ReplaceAllString(s, "$1$3")
		changed = s != s2
		s = s2
	}

	_, err = writer.WriteString(s)
	return err
}

func tagSpacelessParser(doc *Parser, start *Token, arguments *Parser) (INodeTag, error) {
	spacelessNode := &tagSpacelessNode{}

	wrapper, _, err := doc.WrapUntilTag("endspaceless")
	if err != nil {
		return nil, err
	}
	spacelessNode.wrapper = wrapper

	if arguments.Remaining() > 0 {
		return nil, arguments.Error("Malformed spaceless-tag arguments.", nil)
	}

	return spacelessNode, nil
}

func init() {
	mustRegisterTag("spaceless", tagSpacelessParser)
}
