package pongo2

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"sync"
)

// TemplateWriter is the interface used for writing template output.
// It extends io.Writer with a WriteString method for efficient string writing.
// bytes.Buffer satisfies this interface natively.
type TemplateWriter interface {
	io.Writer
	WriteString(string) (int, error)
}

// templateWriter wraps an io.Writer to satisfy the TemplateWriter interface.
// It provides a WriteString implementation for writers that don't have one.
type templateWriter struct {
	w io.Writer
}

// WriteString writes a string to the underlying writer by converting it to bytes.
func (tw *templateWriter) WriteString(s string) (int, error) {
	return tw.w.Write([]byte(s))
}

// Write writes bytes to the underlying writer.
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

	// whitespaceOnce ensures TrimBlocks/LStripBlocks whitespace trimming
	// is applied exactly once, even under concurrent execution.
	whitespaceOnce sync.Once
}

// newTemplateString creates a new template from a byte slice containing template source.
// The template is marked as a string template (not file-based), which affects path
// resolution for include/extends tags. Returns the parsed template or an error.
func newTemplateString(set *TemplateSet, tpl []byte) (*Template, error) {
	return newTemplate(set, "<string>", true, tpl)
}

// newTemplate creates a new template with the given name and source.
// It performs the complete template compilation pipeline:
//  1. Initializes the TemplateSet's built-in tags/filters (once)
//  2. Creates the Template struct with inheritance support structures
//  3. Lexes the source into tokens
//  4. Parses tokens into an AST
//
// Parameters:
//   - set: The TemplateSet this template belongs to
//   - name: Template identifier (file path or "<string>")
//   - isTplString: true if created from string, false if from file
//   - tpl: The raw template source bytes
func newTemplate(set *TemplateSet, name string, isTplString bool, tpl []byte) (*Template, error) {
	strTpl := string(tpl)

	// Mark that a template has been created (prevents further tag/filter banning)
	set.firstTemplateCreated.Store(true)

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

// applyWhitespaceOptions applies TrimBlocks/LStripBlocks whitespace options
// to the template's token list. This is called once at parse time to avoid
// race conditions with concurrent template execution.
//
// Issue #94 https://github.com/flosch/pongo2/issues/94
func (tpl *Template) applyWhitespaceOptions() {
	if !tpl.Options.TrimBlocks && !tpl.Options.LStripBlocks {
		return
	}

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
					t.Val = t.Val[1:]
				}
			}
		}

		prev = t
	}
}

// newContextForExecution prepares the template and context for execution.
// It performs several tasks:
//  1. Walks up the inheritance chain to find the root parent template
//  2. Merges global variables with the provided context
//  3. Validates context keys are valid identifiers
//  4. Checks for naming conflicts between context keys and macros
//
// Returns the root parent template to execute, the execution context, and any error.
func (tpl *Template) newContextForExecution(context Context) (*Template, *ExecutionContext, error) {
	// Apply TrimBlocks/LStripBlocks whitespace options exactly once.
	// Using sync.Once ensures thread-safety for concurrent execution.
	tpl.whitespaceOnce.Do(tpl.applyWhitespaceOptions)

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

// execute is the internal execution method that renders the template to a TemplateWriter.
// It prepares the execution context and runs the root document node's Execute method.
// This is the core execution path used by all public Execute* methods.
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

// newTemplateWriterAndExecute wraps an io.Writer in a templateWriter and executes.
// This allows any io.Writer to be used for template output.
func (tpl *Template) newTemplateWriterAndExecute(context Context, writer io.Writer) error {
	return tpl.execute(context, &templateWriter{w: writer})
}

// newBufferAndExecute creates a pre-sized buffer and executes the template into it.
// The buffer is sized to 130% of the template source size, as templates typically
// expand during rendering (variables, loops, includes, etc.).
// Returns the filled buffer or an error if execution fails.
func (tpl *Template) newBufferAndExecute(context Context) (*bytes.Buffer, error) {
	// Create output buffer. We assume that the rendered template will be 30%
	// larger
	buffer := bytes.NewBuffer(make([]byte, 0, int(float64(tpl.size)*1.3)))
	if err := tpl.execute(context, buffer); err != nil {
		return nil, err
	}
	return buffer, nil
}

// ExecuteWriter executes the template with the given context and writes to writer.
// The output is buffered internally, so nothing is written on error; instead the
// error is returned. This ensures atomic writes - either all output is written
// or none is. Context can be nil for templates that don't require variables.
//
// For high-performance scenarios where partial writes on error are acceptable,
// use ExecuteWriterUnbuffered instead.
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

// ExecuteWriterUnbuffered executes the template and writes directly to the writer
// without intermediate buffering. This provides better performance than ExecuteWriter
// but with a tradeoff: if an error occurs during execution, partial output may have
// already been written to the writer.
//
// Use this method when:
//   - You need maximum performance
//   - You're managing your own buffer pool
//   - Partial output on error is acceptable for your use case
//
// For atomic writes (nothing written on error), use ExecuteWriter instead.
func (tpl *Template) ExecuteWriterUnbuffered(context Context, writer io.Writer) error {
	return tpl.newTemplateWriterAndExecute(context, writer)
}

// ExecuteBytes executes the template and returns the rendered output as a byte slice.
// Context can be nil for templates that don't require variables.
// Returns nil and an error if template execution fails.
func (tpl *Template) ExecuteBytes(context Context) ([]byte, error) {
	// Execute template
	buffer, err := tpl.newBufferAndExecute(context)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// Execute executes the template and returns the rendered output as a string.
// This is the most commonly used execution method for simple use cases.
// Context can be nil for templates that don't require variables.
// Returns an empty string and an error if template execution fails.
func (tpl *Template) Execute(context Context) (string, error) {
	// Execute template
	buffer, err := tpl.newBufferAndExecute(context)
	if err != nil {
		return "", err
	}

	return buffer.String(), nil
}

// ExecuteBlocks executes only the specified named blocks and returns their rendered
// content as a map. This is useful when you need to render specific blocks without
// rendering the entire template, such as for AJAX partial page updates.
//
// Parameters:
//   - context: Variables available during block execution (can be nil)
//   - blocks: List of block names to render
//
// Returns a map where keys are block names and values are their rendered content.
// Blocks not found in the template (or its parents) are omitted from the result.
// The method walks up the template inheritance chain to find all requested blocks.
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
