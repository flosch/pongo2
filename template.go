package pongo2

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

type TemplateWriter interface {
	io.Writer
	WriteString(string) (int, error)
}

type templateWriter struct {
	w io.Writer
}

func (tw *templateWriter) WriteString(s string) (int, error) {
	return tw.w.Write([]byte(s))
}

func (tw *templateWriter) Write(b []byte) (int, error) {
	return tw.w.Write(b)
}

// Template represents a parsed pongo2 template ready for execution.
// It holds the parsed AST (Abstract Syntax Tree) and supports template
// inheritance through parent/child relationships and block overrides.
//
// Templates are created via TemplateSet methods (FromString, FromFile, etc.)
// and should not be instantiated directly. Once parsed, a Template can be
// executed multiple times with different contexts.
//
// The execution flow is:
//
//	Template String → Lexer → Tokens → Parser → AST (root) → Execute → Output
type Template struct {
	// set is the TemplateSet this template belongs to. It provides access to
	// shared configuration, registered tags/filters, template loaders, and
	// global variables. Every template must belong to a TemplateSet.
	set *TemplateSet

	// --- Input fields (set during template creation) ---

	// isTplString indicates whether this template was created from a string
	// (true) or loaded from a file via a TemplateLoader (false). This affects
	// how template paths are resolved for {% include %} and {% extends %} tags.
	isTplString bool

	// name is the identifier for this template. For file-based templates, this
	// is the file path. For string-based templates, this is "<string>". The name
	// is used in error messages and for template caching in the TemplateSet.
	name string

	// size is the length of the template source in bytes. Used to estimate
	// output buffer sizes (templates typically expand ~30% during rendering).
	size int

	// Template inheritance fields: These fields implement Django-style template
	// inheritance. Child templates can extend parent templates and override
	// specific blocks. The inheritance chain is resolved at execution time by
	// walking up the parent chain.
	//
	// IMPORTANT: Block and macro maps use first-come-first-serve semantics.
	// Existing entries must not be overwritten to preserve inheritance
	// behavior.

	// level indicates the depth in the template inheritance hierarchy.
	// The base template has level 0, its child has level 1, and so on.
	// Used for debugging and to prevent infinite inheritance loops.
	level int

	// tokens is the sequence of lexer tokens produced from the template source.
	// These are consumed by the parser to build the AST. Token types include
	// HTML content, variable tags ({{ }}), block tags ({% %}), and comments ({# #}).
	tokens []*Token

	// parent points to the template this one extends (via {% extends %} tag).
	// nil if this is a base template with no parent. During execution,
	// the engine walks up this chain to find the root template to render.
	parent *Template

	// child points to the template that extends this one. This reverse link
	// allows parent templates to delegate block rendering to their children.
	// nil if no template extends this one.
	child *Template

	// blocks maps block names to their NodeWrapper implementations. When a
	// child template defines {% block name %}...{% endblock %}, it registers
	// here. During execution, the most-derived (child-most) block is rendered.
	// Keys are block names, values are the parsed block node wrappers.
	blocks map[string]*NodeWrapper

	// exportedMacros contains macros defined in this template that are available
	// for use in other templates via {% import %}. Only macros explicitly marked
	// for export (or all macros in imported templates) appear here.
	exportedMacros map[string]*tagMacroNode

	// root is the root node of the parsed AST (Abstract Syntax Tree).
	// This nodeDocument contains all parsed template nodes and is the entry
	// point for template execution. Execute() calls root.Execute() to render.
	root *nodeDocument

	// Options allow you to change the behavior of template-engine.
	// You can change the options before calling the Execute method.
	// Includes settings like TrimBlocks and LStripBlocks for whitespace control.
	Options *Options
}

func newTemplateString(set *TemplateSet, tpl []byte) (*Template, error) {
	return newTemplate(set, "<string>", true, tpl)
}

func newTemplate(set *TemplateSet, name string, isTplString bool, tpl []byte) (*Template, error) {
	strTpl := string(tpl)

	// Ensure builtin tags and filters are copied to this template set
	set.initOnce.Do(set.initBuiltins)

	// Create the template
	t := &Template{
		set:            set,
		isTplString:    isTplString,
		name:           name,
		size:           len(strTpl),
		blocks:         make(map[string]*NodeWrapper),
		exportedMacros: make(map[string]*tagMacroNode),
		Options:        newOptions(),
	}
	// Copy all settings from another Options.
	t.Options.Update(set.Options)

	// Tokenize it
	tokens, err := lex(name, strTpl)
	if err != nil {
		return nil, err
	}
	t.tokens = tokens

	// Parse it
	err = t.parse()
	if err != nil {
		return nil, err
	}

	return t, nil
}

func (tpl *Template) newContextForExecution(context Context) (*Template, *ExecutionContext, error) {
	if tpl.Options.TrimBlocks || tpl.Options.LStripBlocks {
		// Issue #94 https://github.com/flosch/pongo2/issues/94
		// If an application configures pongo2 template to trim_blocks,
		// the first newline after a template tag is removed automatically (like in PHP).
		prev := &Token{
			Typ: TokenHTML,
			Val: "\n",
		}

		for _, t := range tpl.tokens {
			if tpl.Options.LStripBlocks {
				if prev.Typ == TokenHTML && t.Typ != TokenHTML && t.Val == "{%" {
					prev.Val = strings.TrimRight(prev.Val, "\t ")
				}
			}

			if tpl.Options.TrimBlocks {
				if prev.Typ != TokenHTML && t.Typ == TokenHTML && prev.Val == "%}" {
					if len(t.Val) > 0 && t.Val[0] == '\n' {
						t.Val = t.Val[1:len(t.Val)]
					}
				}
			}

			prev = t
		}
	}

	// Determine the parent to be executed (for template inheritance)
	parent := tpl
	for parent.parent != nil {
		parent = parent.parent
	}

	// Create context if none is given
	newContext := make(Context)
	newContext.Update(tpl.set.Globals)

	if context != nil {
		newContext.Update(context)

		if len(newContext) > 0 {
			// Check for context name syntax
			err := newContext.checkForValidIdentifiers()
			if err != nil {
				return parent, nil, err
			}

			// Check for clashes with macro names
			for k := range newContext {
				_, has := tpl.exportedMacros[k]
				if has {
					return parent, nil, &Error{
						Filename:  tpl.name,
						Sender:    "execution",
						OrigError: fmt.Errorf("context key name '%s' clashes with macro '%s'", k, k),
					}
				}
			}
		}
	}

	// Create operational context
	ctx := newExecutionContext(parent, newContext)

	return parent, ctx, nil
}

func (tpl *Template) execute(context Context, writer TemplateWriter) error {
	parent, ctx, err := tpl.newContextForExecution(context)
	if err != nil {
		return err
	}

	// Run the selected document
	if err := parent.root.Execute(ctx, writer); err != nil {
		return err
	}

	return nil
}

func (tpl *Template) newTemplateWriterAndExecute(context Context, writer io.Writer) error {
	return tpl.execute(context, &templateWriter{w: writer})
}

func (tpl *Template) newBufferAndExecute(context Context) (*bytes.Buffer, error) {
	// Create output buffer. We assume that the rendered template will be 30%
	// larger
	buffer := bytes.NewBuffer(make([]byte, 0, int(float64(tpl.size)*1.3)))
	if err := tpl.execute(context, buffer); err != nil {
		return nil, err
	}
	return buffer, nil
}

// Executes the template with the given context and writes to writer (io.Writer)
// on success. Context can be nil. Nothing is written on error; instead the error
// is being returned.
func (tpl *Template) ExecuteWriter(context Context, writer io.Writer) error {
	buf, err := tpl.newBufferAndExecute(context)
	if err != nil {
		return err
	}
	_, err = buf.WriteTo(writer)
	if err != nil {
		return err
	}
	return nil
}

// Same as ExecuteWriter. The only difference between both functions is that
// this function might already have written parts of the generated template in the
// case of an execution error because there's no intermediate buffer involved for
// performance reasons. This is handy if you need high performance template
// generation or if you want to manage your own pool of buffers.
func (tpl *Template) ExecuteWriterUnbuffered(context Context, writer io.Writer) error {
	return tpl.newTemplateWriterAndExecute(context, writer)
}

// Executes the template and returns the rendered template as a []byte
func (tpl *Template) ExecuteBytes(context Context) ([]byte, error) {
	// Execute template
	buffer, err := tpl.newBufferAndExecute(context)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// Executes the template and returns the rendered template as a string
func (tpl *Template) Execute(context Context) (string, error) {
	// Execute template
	buffer, err := tpl.newBufferAndExecute(context)
	if err != nil {
		return "", err
	}

	return buffer.String(), nil
}

func (tpl *Template) ExecuteBlocks(context Context, blocks []string) (map[string]string, error) {
	var parents []*Template
	result := make(map[string]string)

	parent := tpl
	for parent != nil {
		// We only want to execute the template if it has a block we want
		for _, block := range blocks {
			if _, ok := tpl.blocks[block]; ok {
				parents = append(parents, parent)
				break
			}
		}
		parent = parent.parent
	}

	for _, t := range parents {
		var buffer *bytes.Buffer
		var ctx *ExecutionContext
		var err error
		for _, blockName := range blocks {
			if _, ok := result[blockName]; ok {
				continue
			}
			if blockWrapper, ok := t.blocks[blockName]; ok {
				// assign the buffer if we haven't done so
				if buffer == nil {
					buffer = bytes.NewBuffer(make([]byte, 0, int(float64(t.size)*1.3)))
				}
				// assign the context if we haven't done so
				if ctx == nil {
					_, ctx, err = t.newContextForExecution(context)
					if err != nil {
						return nil, err
					}
				}
				bErr := blockWrapper.Execute(ctx, buffer)
				if bErr != nil {
					return nil, bErr
				}
				result[blockName] = buffer.String()
				buffer.Reset()
			}
		}
		// We have found all blocks
		if len(blocks) == len(result) {
			break
		}
	}

	return result, nil
}
