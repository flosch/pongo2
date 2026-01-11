package pongo2

import (
	"time"
)

// tagNowNode represents the {% now %} tag.
//
// The now tag outputs the current date and/or time using Go's time formatting.
// The format string uses Go's reference time: Mon Jan 2 15:04:05 MST 2006.
//
// Basic usage:
//
//	{% now "2006-01-02" %}
//
// Output: "2024-03-15" (current date)
//
// Various format examples:
//
//	{% now "January 2, 2006" %}          {# Output: "March 15, 2024" #}
//	{% now "Mon, 02 Jan 2006 15:04:05" %} {# Output: "Fri, 15 Mar 2024 10:30:45" #}
//	{% now "3:04 PM" %}                   {# Output: "10:30 AM" #}
//	{% now "15:04:05" %}                  {# Output: "10:30:45" #}
//	{% now "Monday" %}                    {# Output: "Friday" #}
//
// Common format patterns:
//   - "2006-01-02": ISO date (YYYY-MM-DD)
//   - "01/02/2006": US date format (MM/DD/YYYY)
//   - "02/01/2006": European date format (DD/MM/YYYY)
//   - "15:04:05": 24-hour time
//   - "3:04 PM": 12-hour time with AM/PM
//   - "Mon Jan 2 15:04:05 2006": Full date and time
//
// Using "fake" for testing (outputs fixed date: Feb 5, 2014 18:31:45 UTC):
//
//	{% now "2006-01-02" fake %}
//
// Output: "2014-02-05"
type tagNowNode struct {
	position *Token
	format   string
	fake     bool
}

func (node *tagNowNode) Execute(ctx *ExecutionContext, writer TemplateWriter) error {
	var t time.Time
	if node.fake {
		t = time.Date(2014, time.February, 05, 18, 31, 45, 00, time.UTC)
	} else {
		t = time.Now()
	}

	writer.WriteString(t.Format(node.format))

	return nil
}

func tagNowParser(doc *Parser, start *Token, arguments *Parser) (INodeTag, error) {
	nowNode := &tagNowNode{
		position: start,
	}

	formatToken := arguments.MatchType(TokenString)
	if formatToken == nil {
		return nil, arguments.Error("Expected a format string.", nil)
	}
	nowNode.format = formatToken.Val

	if arguments.MatchOne(TokenIdentifier, "fake") != nil {
		nowNode.fake = true
	}

	if arguments.Remaining() > 0 {
		return nil, arguments.Error("Malformed now-tag arguments.", nil)
	}

	return nowNode, nil
}

func init() {
	RegisterTag("now", tagNowParser)
}
