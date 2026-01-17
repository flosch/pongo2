package pongo2

import (
	"fmt"
	"math/rand"
	"strings"
)

// maxLoremCount limits the maximum number of lorem items to prevent abuse.
const maxLoremCount = 100000

// tagLoremParagraphs and tagLoremWords are pre-split lorem ipsum text.
var (
	tagLoremParagraphs = strings.Split(tagLoremText, "\n")
	tagLoremWords      = strings.Fields(tagLoremText)
)

// tagLoremNode represents the {% lorem %} tag.
//
// The lorem tag generates placeholder "Lorem Ipsum" text, useful for
// prototyping and design mockups.
//
// Basic usage (1 paragraph, plain text):
//
//	{% lorem %}
//
// Output: "Lorem ipsum dolor sit amet, consectetur adipisici elit..."
//
// Specifying number of items:
//
//	{% lorem 3 %}      {# 3 paragraphs #}
//	{% lorem 5 w %}    {# 5 words #}
//	{% lorem 2 p %}    {# 2 HTML paragraphs #}
//
// Method options:
//   - b: Plain text paragraphs (default)
//   - w: Individual words
//   - p: HTML paragraphs wrapped in <p> tags
//
// Using random option (randomizes selection instead of sequential):
//
//	{% lorem 3 p random %}
//
// Examples with output:
//
//	{% lorem 5 w %}
//	{# Output: "Lorem ipsum dolor sit amet" #}
//
//	{% lorem 1 p %}
//	{# Output: "<p>Lorem ipsum dolor sit amet...</p>" #}
//
// Note: The maximum count is 100,000 to prevent abuse.
type tagLoremNode struct {
	position *Token
	count    int    // number of paragraphs
	method   string // w = words, p = HTML paragraphs, b = plain-text (default is b)
	random   bool   // does not use the default paragraph "Lorem ipsum dolor sit amet, ..."
}

// writeLoremItems writes items from the source slice with separator, prefix, and suffix.
func writeLoremItems(writer TemplateWriter, count int, source []string, sep, prefix, suffix string, random bool) error {
	for i := range count {
		if i > 0 {
			if _, err := writer.WriteString(sep); err != nil {
				return err
			}
		}
		if prefix != "" {
			if _, err := writer.WriteString(prefix); err != nil {
				return err
			}
		}
		var item string
		if random {
			item = source[rand.Intn(len(source))] //nolint:gosec // G404: lorem ipsum generation, cryptographic randomness not needed
		} else {
			item = source[i%len(source)]
		}
		if _, err := writer.WriteString(item); err != nil {
			return err
		}
		if suffix != "" {
			if _, err := writer.WriteString(suffix); err != nil {
				return err
			}
		}
	}
	return nil
}

// Execute outputs lorem ipsum text according to the configured method
// (words, paragraphs, or HTML paragraphs) and count.
func (node *tagLoremNode) Execute(ctx *ExecutionContext, writer TemplateWriter) error {
	if node.count > maxLoremCount {
		return ctx.Error(fmt.Sprintf("max count for lorem is %d", maxLoremCount), node.position)
	}

	switch node.method {
	case "b":
		return writeLoremItems(writer, node.count, tagLoremParagraphs, "\n", "", "", node.random)
	case "w":
		return writeLoremItems(writer, node.count, tagLoremWords, " ", "", "", node.random)
	case "p":
		return writeLoremItems(writer, node.count, tagLoremParagraphs, "\n", "<p>", "</p>", node.random)
	default:
		return ctx.OrigError(fmt.Errorf("unsupported method: %s", node.method), nil)
	}
}

// tagLoremParser parses the {% lorem %} tag. It accepts an optional count,
// method (w/p/b), and "random" flag.
func tagLoremParser(doc *Parser, start *Token, arguments *Parser) (INodeTag, error) {
	loremNode := &tagLoremNode{
		position: start,
		count:    1,
		method:   "b",
	}

	if countToken := arguments.MatchType(TokenNumber); countToken != nil {
		loremNode.count = AsValue(countToken.Val).Integer()
	}

	if methodToken := arguments.MatchType(TokenIdentifier); methodToken != nil {
		if methodToken.Val != "w" && methodToken.Val != "p" && methodToken.Val != "b" {
			return nil, arguments.Error("lorem-method must be either 'w', 'p' or 'b'.", nil)
		}

		loremNode.method = methodToken.Val
	}

	if arguments.MatchOne(TokenIdentifier, "random") != nil {
		loremNode.random = true
	}

	if arguments.Remaining() > 0 {
		return nil, arguments.Error("Malformed lorem-tag arguments.", nil)
	}

	return loremNode, nil
}

func init() {
	mustRegisterTag("lorem", tagLoremParser)
}

//nolint:dupword // standard lorem ipsum text naturally contains repeated Latin words
const tagLoremText = `Lorem ipsum dolor sit amet, consectetur adipisici elit, sed eiusmod tempor incidunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquid ex ea commodi consequat. Quis aute iure reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint obcaecat cupiditat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.
Duis autem vel eum iriure dolor in hendrerit in vulputate velit esse molestie consequat, vel illum dolore eu feugiat nulla facilisis at vero eros et accumsan et iusto odio dignissim qui blandit praesent luptatum zzril delenit augue duis dolore te feugait nulla facilisi. Lorem ipsum dolor sit amet, consectetuer adipiscing elit, sed diam nonummy nibh euismod tincidunt ut laoreet dolore magna aliquam erat volutpat.
Ut wisi enim ad minim veniam, quis nostrud exerci tation ullamcorper suscipit lobortis nisl ut aliquip ex ea commodo consequat. Duis autem vel eum iriure dolor in hendrerit in vulputate velit esse molestie consequat, vel illum dolore eu feugiat nulla facilisis at vero eros et accumsan et iusto odio dignissim qui blandit praesent luptatum zzril delenit augue duis dolore te feugait nulla facilisi.
Nam liber tempor cum soluta nobis eleifend option congue nihil imperdiet doming id quod mazim placerat facer possim assum. Lorem ipsum dolor sit amet, consectetuer adipiscing elit, sed diam nonummy nibh euismod tincidunt ut laoreet dolore magna aliquam erat volutpat. Ut wisi enim ad minim veniam, quis nostrud exerci tation ullamcorper suscipit lobortis nisl ut aliquip ex ea commodo consequat.
Duis autem vel eum iriure dolor in hendrerit in vulputate velit esse molestie consequat, vel illum dolore eu feugiat nulla facilisis.
At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. Lorem ipsum dolor sit amet, consetetur sadipscing elitr, At accusam aliquyam diam diam dolore dolores duo eirmod eos erat, et nonumy sed tempor et et invidunt justo labore Stet clita ea et gubergren, kasd magna no rebum. sanctus sea sed takimata ut vero voluptua. est Lorem ipsum dolor sit amet. Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat.
Consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet.`
