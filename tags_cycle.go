package pongo2

// tagCycleValue holds the current value and state of a cycle.
type tagCycleValue struct {
	node  *tagCycleNode
	value *Value
}

// tagCycleNode represents the {% cycle %} tag.
//
// The cycle tag cycles through a list of values each time it is encountered.
// It's commonly used within loops to alternate between values (e.g., alternating
// row colors in a table).
//
// Basic usage (cycles through values on each iteration):
//
//	{% for item in items %}
//	    <tr class="{% cycle 'odd' 'even' %}">
//	        <td>{{ item }}</td>
//	    </tr>
//	{% endfor %}
//
// Output (for 4 items):
//
//	<tr class="odd"><td>...</td></tr>
//	<tr class="even"><td>...</td></tr>
//	<tr class="odd"><td>...</td></tr>
//	<tr class="even"><td>...</td></tr>
//
// Using "as" to store the cycle value in a variable:
//
//	{% cycle 'red' 'green' 'blue' as color %}
//	<p style="color: {{ color }}">Text</p>
//	{% cycle color %}  {# Advances to next value #}
//	<p style="color: {{ color }}">More text</p>
//
// Using "silent" to not output the value (only store it):
//
//	{% cycle 'a' 'b' 'c' as letter silent %}
//	Current letter: {{ letter }}
type tagCycleNode struct {
	position *Token
	args     []IEvaluator
	asName   string
	silent   bool
}

// String returns the string representation of the current cycle value.
func (cv *tagCycleValue) String() string {
	return cv.value.String()
}

// cycleIdx returns the current cycle index for this node from the execution
// context and advances it. Each template execution gets its own independent
// cycle state via ctx.tagState.
func (node *tagCycleNode) cycleIdx(ctx *ExecutionContext) int {
	idx, _ := ctx.tagState[node].(int)
	ctx.tagState[node] = idx + 1
	return idx
}

// Execute outputs the next value in the cycle sequence. If the cycle was
// stored with "as", it updates the stored value and optionally outputs it
// (unless "silent" was specified).
func (node *tagCycleNode) Execute(ctx *ExecutionContext, writer TemplateWriter) error {
	idx := node.cycleIdx(ctx)
	item := node.args[idx%len(node.args)]

	val, err := item.Evaluate(ctx)
	if err != nil {
		return err
	}

	if t, ok := val.Interface().(*tagCycleValue); ok {
		// {% cycle cycleitem %} — advance the referenced cycle node
		refIdx := t.node.cycleIdx(ctx)
		item := t.node.args[refIdx%len(t.node.args)]

		val, err := item.Evaluate(ctx)
		if err != nil {
			return err
		}

		t.value = val

		if !t.node.silent {
			if _, err := writer.WriteString(val.String()); err != nil {
				return err
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
				return err
			}
		}
	}

	return nil
}

// tagCycleParser parses the {% cycle %} tag. It accepts multiple values
// to cycle through, with optional "as name" to store the cycle and "silent"
// to suppress output.
// HINT: We're not supporting the old comma-separated list of expressions argument-style
func tagCycleParser(doc *Parser, start *Token, arguments *Parser) (INodeTag, error) {
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
				return nil, arguments.Error("Name (identifier) expected after 'as'.", nil)
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
		return nil, arguments.Error("Malformed cycle-tag.", nil)
	}

	if len(cycleNode.args) == 0 {
		return nil, arguments.Error("'cycle' tag requires at least one argument.", nil)
	}

	return cycleNode, nil
}

func init() {
	mustRegisterTag("cycle", tagCycleParser)
}
