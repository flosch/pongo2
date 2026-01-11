package pongo2

// tagIfNode represents the {% if %} tag.
//
// The if tag evaluates a condition and renders its contents only if the condition
// is true. It supports elif (else if) and else clauses for complex conditional logic.
//
// Basic usage:
//
//	{% if condition %}
//	    Content shown when condition is true.
//	{% endif %}
//
// Using else:
//
//	{% if user.is_authenticated %}
//	    Welcome back, {{ user.name }}!
//	{% else %}
//	    Please log in.
//	{% endif %}
//
// Using elif (else if):
//
//	{% if score >= 90 %}
//	    Grade: A
//	{% elif score >= 80 %}
//	    Grade: B
//	{% elif score >= 70 %}
//	    Grade: C
//	{% else %}
//	    Grade: F
//	{% endif %}
//
// Supported operators in conditions:
//   - Comparison: ==, !=, <, >, <=, >=
//   - Logical: and, or, not
//   - Membership: in
//
// Examples with operators:
//
//	{% if user.age >= 18 and user.country == "US" %}
//	    Adult US user.
//	{% endif %}
//
//	{% if "admin" in user.roles %}
//	    Admin panel link.
//	{% endif %}
//
//	{% if not user.is_banned %}
//	    User can post.
//	{% endif %}
type tagIfNode struct {
	conditions []IEvaluator
	wrappers   []*NodeWrapper
}

func (node *tagIfNode) Execute(ctx *ExecutionContext, writer TemplateWriter) error {
	for i, condition := range node.conditions {
		result, err := condition.Evaluate(ctx)
		if err != nil {
			return err
		}

		if result.IsTrue() {
			return node.wrappers[i].Execute(ctx, writer)
		}
		// Last condition?
		if len(node.conditions) == i+1 && len(node.wrappers) > i+1 {
			return node.wrappers[i+1].Execute(ctx, writer)
		}
	}
	return nil
}

func tagIfParser(doc *Parser, start *Token, arguments *Parser) (INodeTag, error) {
	ifNode := &tagIfNode{}

	// Parse first and main IF condition
	condition, err := arguments.ParseExpression()
	if err != nil {
		return nil, err
	}
	ifNode.conditions = append(ifNode.conditions, condition)

	if arguments.Remaining() > 0 {
		return nil, arguments.Error("If-condition is malformed.", nil)
	}

	// Check the rest
	for {
		wrapper, tagArgs, err := doc.WrapUntilTag("elif", "else", "endif")
		if err != nil {
			return nil, err
		}
		ifNode.wrappers = append(ifNode.wrappers, wrapper)

		if wrapper.Endtag == "elif" {
			// elif can take a condition
			condition, err = tagArgs.ParseExpression()
			if err != nil {
				return nil, err
			}
			ifNode.conditions = append(ifNode.conditions, condition)

			if tagArgs.Remaining() > 0 {
				return nil, tagArgs.Error("Elif-condition is malformed.", nil)
			}
		} else {
			if tagArgs.Count() > 0 {
				// else/endif can't take any conditions
				return nil, tagArgs.Error("Arguments not allowed here.", nil)
			}
		}

		if wrapper.Endtag == "endif" {
			break
		}
	}

	return ifNode, nil
}

func init() {
	RegisterTag("if", tagIfParser)
}
