// Package pongo2 implements a Django-syntax template engine for Go.
package pongo2

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

const (
	// EOF represents the end-of-file marker used by the lexer to signal
	// that all input has been consumed. The value -1 is chosen because
	// it's an invalid rune value that cannot appear in valid UTF-8 input.
	EOF rune = -1
)

// TokenType represents the classification of a lexer token.
// Each token produced by the lexer is assigned one of these types
// to indicate what kind of template element it represents.
type TokenType int

const (
	// TokenError indicates a lexical error was encountered. The token's
	// Val field contains the error message describing what went wrong.
	TokenError TokenType = iota

	// TokenHTML represents raw HTML/text content outside of template tags.
	// This is any content not enclosed in {{ }}, {% %}, or {# #} delimiters.
	TokenHTML

	// TokenKeyword represents a reserved word in the template language.
	// Keywords include: in, and, or, not, true, false, as, export.
	TokenKeyword

	// TokenIdentifier represents a variable name, filter name, or tag name.
	// Identifiers start with a letter or underscore, followed by letters,
	// digits, or underscores (e.g., "user", "my_var", "_private").
	TokenIdentifier

	// TokenString represents a quoted string literal.
	// Strings can be enclosed in single or double quotes and support
	// escape sequences: \\, \", \', \n, \t, \r.
	TokenString

	// TokenNumber represents a numeric literal (integers only).
	// The lexer currently only supports integer literals; floating-point
	// numbers are handled contextually during parsing.
	TokenNumber

	// TokenSymbol represents an operator or punctuation symbol.
	// Symbols include: {{, }}, {%, %}, ==, !=, <=, >=, +, -, etc.
	TokenSymbol

	// TokenNil is reserved but currently unused.
	// FIXME: It seems like TokenNil is never used as a token. Either remove
	// TokenNil entirely or use it.
	TokenNil
)

var (
	// tokenSpaceChars defines whitespace characters that are skipped
	// when tokenizing inside template tags ({{ }} and {% %}).
	tokenSpaceChars = " \n\r\t"

	// tokenIdentifierChars defines valid starting characters for identifiers.
	// Identifiers must begin with a letter (a-z, A-Z) or underscore.
	tokenIdentifierChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_"

	// tokenIdentifierCharsWithDigits defines valid continuation characters
	// for identifiers. After the first character, digits are also allowed.
	tokenIdentifierCharsWithDigits = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_0123456789"

	// tokenDigits defines valid digit characters for number literals.
	tokenDigits = "0123456789"

	// TokenSymbols lists all recognized operator and punctuation symbols.
	// The slice is ordered by symbol length (longest first) to ensure
	// greedy matching: "{{-" is matched before "{{".
	//
	// Symbol categories:
	//   - 3-char: Whitespace-trimming delimiters ({{-, -}}, {%-, -%})
	//   - 2-char: Comparison ops, logical ops, template delimiters
	//   - 1-char: Arithmetic, punctuation, filter pipe, etc.
	TokenSymbols = []string{
		// 3-Char symbols
		"{{-", "-}}", "{%-", "-%}",

		// 2-Char symbols
		"==", ">=", "<=", "&&", "||", "{{", "}}", "{%", "%}", "!=", "<>",

		// 1-Char symbol
		"(", ")", "+", "-", "*", "<", ">", "/", "^", ",", ".", "!", "|", ":", "=", "%", "[", "]",
	}

	// TokenKeywords lists all reserved words in the template language.
	// These cannot be used as variable or filter names.
	TokenKeywords = []string{"in", "and", "or", "not", "true", "false", "as", "export"}

	// tokenKeywordsMap is a pre-compiled map for O(1) keyword lookup.
	// This is more efficient than iterating through the TokenKeywords slice.
	tokenKeywordsMap = map[string]struct{}{
		"in":     {},
		"and":    {},
		"or":     {},
		"not":    {},
		"true":   {},
		"false":  {},
		"as":     {},
		"export": {},
	}

	// stringEscapeReplacer is a pre-compiled replacer for handling escape
	// sequences in string tokens. Using a package-level variable avoids
	// creating a new Replacer for each string token.
	stringEscapeReplacer = strings.NewReplacer(
		`\\`, `\`,
		`\"`, `"`,
		`\'`, `'`,
		`\n`, "\n",
		`\t`, "\t",
		`\r`, "\r",
	)
)

// Token represents a single lexical element produced by the lexer.
// Tokens are the output of lexical analysis and the input to the parser.
type Token struct {
	// Filename is the name of the template file this token came from.
	// Used for error reporting to help users locate issues.
	Filename string

	// Typ indicates what kind of token this is (HTML, identifier, etc.).
	Typ TokenType

	// Val contains the actual text content of the token.
	// For TokenError, this contains the error message.
	Val string

	// Line is the 1-based line number where this token starts.
	Line int

	// Col is the 1-based column number where this token starts.
	Col int

	// TrimWhitespaces is true for whitespace-trimming delimiters ({{-, -}}, {%-, -%}).
	// When true, adjacent whitespace in HTML content should be stripped.
	TrimWhitespaces bool
}

// lexerStateFn represents a state function in the lexer's state machine.
// Each state function processes input and returns the next state to enter,
// or nil to terminate lexing. This pattern enables clean separation of
// lexing logic for different token types (strings, numbers, identifiers, etc.).
type lexerStateFn func() lexerStateFn

// lexer implements a state-machine based tokenizer for pongo2 templates.
// It scans the input string character by character, identifying template
// constructs and emitting tokens for the parser to consume.
//
// The lexer handles:
//   - Raw HTML content (outside template tags)
//   - Variable tags: {{ expression }}
//   - Block tags: {% tag_name ... %}
//   - Comments: {# comment #}
//   - Verbatim blocks: {% verbatim %}...{% endverbatim %}
//   - Whitespace trimming: {{- ... -}}, {%- ... -%}
type lexer struct {
	// name is the template name/filename for error reporting.
	name string

	// input is the complete template source being lexed.
	input string

	// start is the byte position where the current token begins.
	start int

	// pos is the current byte position in the input (cursor).
	pos int

	// width is the byte width of the last rune read by next().
	// Used by backup() to step back one rune.
	width int

	// tokens accumulates all tokens produced during lexing.
	tokens []*Token

	// errored is set to true when a lexical error occurs.
	// The error details are stored in the last token (TokenError).
	errored bool

	// startline is the line number where the current token begins.
	startline int

	// startcol is the column number where the current token begins.
	startcol int

	// line is the current line number (1-based) in the input.
	line int

	// col is the current column number (1-based) in the input.
	col int

	// inVerbatim is true when inside a {% verbatim %} block.
	// In verbatim mode, template tags are treated as raw HTML.
	inVerbatim bool
}

// String returns a human-readable representation of the token for debugging.
// Long values (>1000 chars) are truncated to show the beginning and end.
// Format: <Token Typ=TYPE (num) Val='value' Line=N Col=N, WT=bool>
func (t *Token) String() string {
	val := t.Val
	if len(val) > 1000 {
		val = fmt.Sprintf("%s...%s", val[:10], val[len(val)-5:])
	}

	typ := ""
	switch t.Typ {
	case TokenHTML:
		typ = "HTML"
	case TokenError:
		typ = "Error"
	case TokenIdentifier:
		typ = "Identifier"
	case TokenKeyword:
		typ = "Keyword"
	case TokenNumber:
		typ = "Number"
	case TokenString:
		typ = "String"
	case TokenSymbol:
		typ = "Symbol"
	case TokenNil:
		typ = "Nil"
	default:
		typ = "Unknown"
	}

	return fmt.Sprintf("<Token Typ=%s (%d) Val='%s' Line=%d Col=%d, WT=%t>",
		typ, t.Typ, val, t.Line, t.Col, t.TrimWhitespaces)
}

// lex tokenizes the given template source string and returns a slice of tokens.
// This is the main entry point for lexical analysis.
//
// Parameters:
//   - name: The template name/filename (used for error messages)
//   - input: The complete template source to tokenize
//
// Returns the token slice on success, or an Error with location info on failure.
func lex(name string, input string) ([]*Token, error) {
	l := &lexer{
		name:      name,
		input:     input,
		tokens:    make([]*Token, 0, 100),
		line:      1,
		col:       1,
		startline: 1,
		startcol:  1,
	}
	l.run()
	if l.errored {
		errtoken := l.tokens[len(l.tokens)-1]
		return nil, &Error{
			Filename:  name,
			Line:      errtoken.Line,
			Column:    errtoken.Col,
			Sender:    "lexer",
			OrigError: errors.New(errtoken.Val),
		}
	}
	return l.tokens, nil
}

// value returns the substring of input from start to current position.
// This is the text content of the token currently being built.
func (l *lexer) value() string {
	return l.input[l.start:l.pos]
}

// length returns the byte length of the current token being built.
func (l *lexer) length() int {
	return l.pos - l.start
}

// emit creates a token of the given type from the current lexer state
// and appends it to the token list. After emitting, the lexer's start
// position is advanced to prepare for the next token.
//
// Special handling:
//   - TokenString: Escape sequences are processed (\\, \", \', \n, \t, \r)
//   - TokenSymbol: Whitespace-trimming symbols ({{-, -}}, {%-, -%}) have
//     TrimWhitespaces set to true and the "-" is removed from Val
func (l *lexer) emit(t TokenType) {
	tok := &Token{
		Filename: l.name,
		Typ:      t,
		Val:      l.value(),
		Line:     l.startline,
		Col:      l.startcol,
	}

	if t == TokenString {
		// Escape sequences in strings
		tok.Val = stringEscapeReplacer.Replace(tok.Val)
	}

	if t == TokenSymbol && len(tok.Val) == 3 && (strings.HasSuffix(tok.Val, "-") || strings.HasPrefix(tok.Val, "-")) {
		tok.TrimWhitespaces = true
		tok.Val = strings.ReplaceAll(tok.Val, "-", "")
	}

	l.tokens = append(l.tokens, tok)
	l.start = l.pos
	l.startline = l.line
	l.startcol = l.col
}

// next advances the lexer by one rune and returns it.
// Returns EOF if the end of input has been reached.
// Updates pos and col to reflect the new position.
// The width of the rune is stored for use by backup().
func (l *lexer) next() rune {
	if l.pos >= len(l.input) {
		l.width = 0
		return EOF
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = w
	l.pos += l.width
	l.col++
	return r
}

// backup steps back one rune in the input.
// Can only be called once per call to next().
// Used to "unread" a character after peeking or when a character
// doesn't match expected input.
func (l *lexer) backup() {
	l.pos -= l.width
	l.col--
}

// peek returns the next rune without consuming it.
// Equivalent to calling next() followed by backup().
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// ignore discards the text from start to the current position.
// Used to skip over content that shouldn't become a token (e.g., comments).
// After ignore(), the next emit() will start from the current position.
func (l *lexer) ignore() {
	l.start = l.pos
	l.startline = l.line
	l.startcol = l.col
}

// accept consumes the next rune if it's contained in the valid string.
// Returns true if a rune was consumed, false otherwise.
// If false, the lexer position is unchanged.
func (l *lexer) accept(what string) bool {
	if strings.ContainsRune(what, l.next()) {
		return true
	}
	l.backup()
	return false
}

// acceptRun consumes a run of runes from the valid set.
// Continues consuming as long as each rune is in the valid string.
// Stops (and backs up) when a non-matching rune is encountered.
func (l *lexer) acceptRun(what string) {
	for strings.ContainsRune(what, l.next()) {
	}
	l.backup()
}

// errorf records a lexical error and terminates the current state.
// Creates a TokenError with the formatted message and sets the errored flag.
// Always returns nil to signal that lexing should stop.
func (l *lexer) errorf(format string, args ...any) lexerStateFn {
	t := &Token{
		Filename: l.name,
		Typ:      TokenError,
		Val:      fmt.Sprintf(format, args...),
		Line:     l.startline,
		Col:      l.startcol,
	}
	l.tokens = append(l.tokens, t)
	l.errored = true
	l.startline = l.line
	l.startcol = l.col
	return nil
}

// emitRemainingHTML emits any accumulated HTML content as a TokenHTML.
// Called before entering a template tag or at end of input to flush
// any raw HTML that was being collected.
func (l *lexer) emitRemainingHTML() {
	if l.pos > l.start {
		l.emit(TokenHTML)
	}
}

// ignoreSingleLineComment skips over a single-line comment {# ... #}.
// Comments are not emitted as tokens; they are completely discarded.
// Reports an error if the comment is not closed or contains a newline.
func (l *lexer) ignoreSingleLineComment() {
	if !strings.HasPrefix(l.input[l.pos:], "{#") {
		return
	}

	l.emitRemainingHTML()

	l.pos += 2 // pass '{#'
	l.col += 2

	for {
		switch l.peek() {
		case EOF:
			l.errorf("Single-line comment not closed.")
			return
		case '\n':
			l.errorf("Newline not permitted in a single-line comment.")
			return
		}

		if strings.HasPrefix(l.input[l.pos:], "#}") {
			l.pos += 2 // pass '#}'
			l.col += 2
			break
		}

		l.next()
	}
	l.ignore() // ignore whole comment
}

// processVerbatimTag handles {% verbatim %} and {% endverbatim %} tags.
// Content inside verbatim blocks is treated as raw HTML, not parsed as
// template syntax. This allows including literal {{ }} or {% %} in output.
//
// TODO: Support verbatim tag names as per Django docs:
// https://docs.djangoproject.com/en/dev/ref/templates/builtins/#verbatim
func (l *lexer) processVerbatimTag() {
	if l.inVerbatim {
		// end verbatim
		if strings.HasPrefix(l.input[l.pos:], "{% endverbatim %}") {
			l.emitRemainingHTML()
			w := len("{% endverbatim %}")
			l.pos += w
			l.col += w
			l.ignore()
			l.inVerbatim = false
		}
	} else if strings.HasPrefix(l.input[l.pos:], "{% verbatim %}") { // tag
		l.emitRemainingHTML()
		l.inVerbatim = true
		w := len("{% verbatim %}")
		l.pos += w
		l.col += w
		l.ignore()
	}
}

// run is the main lexer loop that processes the entire input.
// It iterates through the input, handling verbatim blocks, comments,
// and template tags ({{ }} and {% %}). Raw HTML content between tags
// is accumulated and emitted as TokenHTML.
//
// The loop terminates when EOF is reached or an error occurs.
func (l *lexer) run() {
	for {
		l.processVerbatimTag()

		if !l.inVerbatim {
			// Ignore single-line comments {# ... #}
			l.ignoreSingleLineComment()
			if l.errored {
				return
			}

			if strings.HasPrefix(l.input[l.pos:], "{{") || // variable
				strings.HasPrefix(l.input[l.pos:], "{%") { // tag
				l.emitRemainingHTML()
				l.tokenizeTemplateCode()
				if l.errored {
					return
				}
				continue
			}
		}

		// Advance line and reset column upon new line.
		switch l.peek() {
		case '\n':
			l.line++
			l.col = 0
		}

		// Stop lexing once EOF is reached.
		if l.next() == EOF {
			break
		}
	}

	l.emitRemainingHTML()

	if l.inVerbatim {
		l.errorf("verbatim-tag not closed, got EOF.")
	}
}

// tokenizeTemplateCode runs the state machine to process a single template variable/tag.
// Called when {{ or {% is encountered to tokenizeTemplateCode the contents. Starts in
// stateCode and continues until a terminal state (nil) is reached.
func (l *lexer) tokenizeTemplateCode() {
	for state := l.stateCode; state != nil; {
		state = state()
	}
}

// stateCode is the main state for tokenizing inside template tags.
// It handles whitespace, identifies the start of identifiers, numbers,
// strings, and symbols, and dispatches to the appropriate sub-state.
//
// Returns nil when a closing delimiter (}}, %}. -}}, -%}) is encountered.
func (l *lexer) stateCode() lexerStateFn {
outer_loop:
	for {
		switch {
		case l.accept(tokenSpaceChars):
			if l.value() == "\n" {
				return l.errorf("Newline not allowed within tag/variable.")
			}
			l.ignore()
			continue
		case l.accept(tokenIdentifierChars):
			return l.stateIdentifier
		case l.accept(tokenDigits):
			return l.stateNumber
		case l.accept(`"'`):
			return l.stateString
		}

		// Check for symbol
		for _, sym := range TokenSymbols {
			if strings.HasPrefix(l.input[l.start:], sym) {
				l.pos += len(sym)
				l.col += l.length()
				l.emit(TokenSymbol)

				if sym == "%}" || sym == "-%}" || sym == "}}" || sym == "-}}" {
					// Tag/variable end, return after emit
					return nil
				}

				continue outer_loop
			}
		}

		break
	}

	// Normal shut down
	return nil
}

// stateIdentifier lexes an identifier or keyword token.
// Called after the first identifier character has been accepted.
// Consumes remaining identifier characters, then checks if the result
// is a keyword (emits TokenKeyword) or regular identifier (emits TokenIdentifier).
func (l *lexer) stateIdentifier() lexerStateFn {
	l.acceptRun(tokenIdentifierChars)
	l.acceptRun(tokenIdentifierCharsWithDigits)
	val := l.value()
	if _, isKeyword := tokenKeywordsMap[val]; isKeyword {
		l.emit(TokenKeyword)
		return l.stateCode
	}
	l.emit(TokenIdentifier)
	return l.stateCode
}

// stateNumber lexes a numeric literal token.
// Called after the first digit has been accepted.
// Currently only handles integers. If an identifier character follows
// the digits, treats the whole thing as an identifier (see issue #151).
//
// Note: Floating-point numbers are not directly supported in the lexer.
// Expressions like "user.0" use the dot as an access operator, and
// floating-point comparisons like "score >= 8.5" would need context-sensitive
// parsing to distinguish from property access.
func (l *lexer) stateNumber() lexerStateFn {
	l.acceptRun(tokenDigits)
	if l.accept(tokenIdentifierCharsWithDigits) {
		// This seems to be an identifier starting with a number.
		// See https://github.com/flosch/pongo2/issues/151
		return l.stateIdentifier()
	}
	/*
		Maybe context-sensitive number lexing?
		* comments.0.Text // first comment
		* usercomments.1.0 // second user, first comment
		* if (score >= 8.5) // 8.5 as a number

		if l.peek() == '.' {
			l.accept(".")
			if !l.accept(tokenDigits) {
				return l.errorf("Malformed number.")
			}
			l.acceptRun(tokenDigits)
		}
	*/
	l.emit(TokenNumber)
	return l.stateCode
}

// stateString lexes a quoted string literal.
// Called after the opening quote (single or double) has been accepted.
// Handles escape sequences: \\, \", \', \n, \t, \r.
// Reports errors for unclosed strings, unknown escapes, or embedded newlines.
func (l *lexer) stateString() lexerStateFn {
	quotationMark := l.value()
	l.ignore()
	l.startcol-- // we're starting the position at the first "
	for !l.accept(quotationMark) {
		switch l.next() {
		case '\\':
			// escape sequence
			switch l.peek() {
			case '"', '\'', '\\', 'n', 't', 'r':
				l.next()
			default:
				return l.errorf("Unknown escape sequence: \\%c", l.peek())
			}
		case EOF:
			return l.errorf("Unexpected EOF, string not closed.")
		case '\n':
			return l.errorf("Newline in string is not allowed.")
		}
	}
	l.backup()
	l.emit(TokenString)

	l.next()
	l.ignore()

	return l.stateCode
}
