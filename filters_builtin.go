package pongo2

/* Filters that won't be added:
   ----------------------------

   get_static_prefix (reason: web-framework specific)
   pprint (reason: python-specific)
   static (reason: web-framework specific)
*/

/*
   Notable Django behavior references for filter implementations:
   ---------------------------------------------------------------

   - title: Django uses a custom algorithm that capitalizes after any non-letter/digit
     character, then fixes apostrophe and digit cases with regex.
     Django ref: django/utils/text.py capfirst(), re.sub for apostrophe/digit.

   - center: Padding bias for odd margins follows Python's str.center():
     left = marg/2 + (marg & width & 1). Both odd → extra on left, else right.
     Django ref: django/template/defaultfilters.py center() → str.center().

   - slugify: Django preserves underscores in slugs. Underscores are NOT stripped.
     Django ref: django/utils/text.py slugify().

   - filesizeformat: Uses non-breaking space (U+00A0) between number and unit,
     singular "byte" (not "bytes") for exactly 1 byte, and "-" prefix for negatives.
     Django ref: django/template/defaultfilters.py filesizeformat().

   - timesince/timeuntil: Uses calendar-based month/year arithmetic (not fixed
     365/30 approximations). Shows only adjacent time units. Returns "0 minutes"
     for reversed dates.
     Django ref: django/utils/timesince.py.

   - linebreaks: Groups text into paragraphs split by double newlines. Within
     paragraphs, single newlines become <br />. The paragraph algorithm is:
     split by 2+ newlines → each paragraph gets <p>...</p>.
     Django ref: django/utils/html.py linebreaks().

   - unordered_list: Uses tab indentation with depth starting at 1. Items are
     separated by newlines. Nested sublists get <ul>/<li> on separate lines.
     Django ref: django/template/defaultfilters.py unordered_list() → list_formatter().

   - escapejs: Escapes characters 0x00-0x1F, \, ', ", `, <, >, &, =, -, ;,
     U+2028, U+2029 to \uXXXX format (uppercase hex in Django, lowercase in Go).
     Django ref: django/utils/html.py _js_escapes table.

   - wordwrap: Wraps at character column width (not word count). Uses word-boundary
     breaking (long words are not split). Preserves existing newlines.
     Normalizes \r\n and \r to \n before processing. Verified against Django 4.2.
     Django ref: django/utils/text.py wrap().

   - floatformat: Negative arg means "display N decimal places unless the result
     would be all zeros." Positive arg always shows exactly N places.
     Django ref: django/template/defaultfilters.py floatformat().

   - forloop: Django's forloop has counter, counter0, revcounter, revcounter0,
     first, last, parentloop — but NOT a .length attribute.
     Django ref: django/template/defaulttags.py ForNode.

   - urlencode: Uses Go's url.QueryEscape. Django difference: Django uses
     urllib.parse.quote(safe='/') which encodes spaces as %20 and preserves /.
     Go's url.QueryEscape encodes spaces as + and encodes / as %2F.

   - iriencode: Uses Go's url.QueryEscape for non-IRI characters. Django
     difference: Django's iri_to_uri() uses urllib.parse.quote() which encodes
     spaces as %20. Go's url.QueryEscape encodes spaces as +.

   - truncatewords: Verified against Django 4.2 with script. Matches Django's
     Truncator(value).words(length, truncate=" …") — space before ellipsis is
     intentional.

   - truncatechars: Verified against Django 4.2 with script.

   - truncatechars_html / truncatewords_html: Returns AsSafeValue() to prevent
     double-escaping of preserved HTML tags. Django difference: Django uses
     is_safe=True which preserves input safety status; pongo2 unconditionally
     marks output as safe. This is more user-friendly for HTML-producing filters.

   - escapejs: Verified against Django 4.2 with script. Uses lowercase hex
     (\u000d) instead of Django's uppercase (\u000D) — functionally equivalent.

   - phone2numeric: Verified against Django 4.2 with script.

   - divisibleby: Verified against Django 4.2 with script.

   - yesno: Verified against Django 4.2 with script.

   - templatetag: Verified against Django 4.2 with script.

   Intentional differences from Django:
   - stringformat: Uses Go fmt format verbs instead of Python % formatting.
   - date: Uses Go time formatting (reference time: Mon Jan 2 15:04:05 MST 2006)
     instead of Django's PHP-style format characters.
   - now tag: Uses Go time formatting instead of Django format characters.
   - lorem tag: Uses static pre-defined paragraphs; Django generates random text
     from a word list.
   - firstof tag: Does not support "as variable_name" syntax (feature gap).
*/

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

func mustRegisterFilter(name string, fn FilterFunction) {
	if err := registerFilterBuiltin(name, fn); err != nil {
		panic(err)
	}
}

// htmlEscapeReplacer is a pre-compiled replacer for HTML escaping.
// Using a single Replacer is more efficient than multiple strings.Replace calls
// because it processes the string in a single pass.
var htmlEscapeReplacer = strings.NewReplacer(
	"&", "&amp;",
	">", "&gt;",
	"<", "&lt;",
	`"`, "&quot;",
	"'", "&#39;",
)

// stripTagsIteratively applies tag-stripping regex patterns iteratively until convergence.
// This handles obfuscated tags like "<sc<script>ript>" which become "<script>" after first pass.
// Returns an error if stripping doesn't converge within maxIterations.
func stripTagsIteratively(s string, patterns []*regexp.Regexp, maxIterations int, filterName string) (string, error) {
	for i := range maxIterations {
		prev := s
		for _, re := range patterns {
			s = re.ReplaceAllString(s, "")
		}
		if s == prev {
			return strings.TrimSpace(s), nil
		}
		if i == maxIterations-1 {
			return "", &Error{
				Sender:    filterName,
				OrigError: fmt.Errorf("tag stripping did not converge after max iterations (%d); input may be maliciously crafted", maxIterations),
			}
		}
	}
	return strings.TrimSpace(s), nil
}

// addslashesReplacer is a pre-compiled replacer for adding slashes.
var addslashesReplacer = strings.NewReplacer(
	`\`, `\\`,
	`"`, `\"`,
	"'", `\'`,
)

func init() {
	mustRegisterFilter("escape", filterEscape)
	mustRegisterFilter("e", filterEscape) // alias of `escape`
	mustRegisterFilter("safe", filterSafe)
	mustRegisterFilter("escapejs", filterEscapejs)

	mustRegisterFilter("add", filterAdd)
	mustRegisterFilter("addslashes", filterAddslashes)
	mustRegisterFilter("capfirst", filterCapfirst)
	mustRegisterFilter("center", filterCenter)
	mustRegisterFilter("cut", filterCut)
	mustRegisterFilter("date", filterDate)
	mustRegisterFilter("default", filterDefault)
	mustRegisterFilter("default_if_none", filterDefaultIfNone)
	mustRegisterFilter("divisibleby", filterDivisibleby)
	mustRegisterFilter("first", filterFirst)
	mustRegisterFilter("floatformat", filterFloatformat)
	mustRegisterFilter("get_digit", filterGetdigit)
	mustRegisterFilter("iriencode", filterIriencode)
	mustRegisterFilter("join", filterJoin)
	mustRegisterFilter("last", filterLast)
	mustRegisterFilter("length", filterLength)
	mustRegisterFilter("length_is", filterLengthis)
	mustRegisterFilter("linebreaks", filterLinebreaks)
	mustRegisterFilter("linebreaksbr", filterLinebreaksbr)
	mustRegisterFilter("linenumbers", filterLinenumbers)
	mustRegisterFilter("ljust", filterLjust)
	mustRegisterFilter("lower", filterLower)
	mustRegisterFilter("make_list", filterMakelist)
	mustRegisterFilter("phone2numeric", filterPhone2numeric)
	mustRegisterFilter("pluralize", filterPluralize)
	mustRegisterFilter("random", filterRandom)
	mustRegisterFilter("removetags", filterRemovetags)
	mustRegisterFilter("rjust", filterRjust)
	mustRegisterFilter("slice", filterSlice)
	mustRegisterFilter("split", filterSplit)
	mustRegisterFilter("stringformat", filterStringformat)
	mustRegisterFilter("striptags", filterStriptags)
	mustRegisterFilter("time", filterDate) // time uses filterDate (same golang-format)
	mustRegisterFilter("title", filterTitle)
	mustRegisterFilter("truncatechars", filterTruncatechars)
	mustRegisterFilter("truncatechars_html", filterTruncatecharsHTML)
	mustRegisterFilter("truncatewords", filterTruncatewords)
	mustRegisterFilter("truncatewords_html", filterTruncatewordsHTML)
	mustRegisterFilter("upper", filterUpper)
	mustRegisterFilter("urlencode", filterUrlencode)
	mustRegisterFilter("urlize", filterUrlize)
	mustRegisterFilter("urlizetrunc", filterUrlizetrunc)
	mustRegisterFilter("wordcount", filterWordcount)
	mustRegisterFilter("wordwrap", filterWordwrap)
	mustRegisterFilter("yesno", filterYesno)
	mustRegisterFilter("timesince", filterTimesince)
	mustRegisterFilter("timeuntil", filterTimeuntil)
	mustRegisterFilter("dictsort", filterDictsort)
	mustRegisterFilter("dictsortreversed", filterDictsortReversed)
	mustRegisterFilter("unordered_list", filterUnorderedList)
	mustRegisterFilter("slugify", filterSlugify)
	mustRegisterFilter("filesizeformat", filterFilesizeformat)
	mustRegisterFilter("safeseq", filterSafeseq)
	mustRegisterFilter("escapeseq", filterEscapeseq)
	mustRegisterFilter("json_script", filterJSONScript)

	mustRegisterFilter("float", filterFloat)     // pongo-specific
	mustRegisterFilter("integer", filterInteger) // pongo-specific
}

const ellipsis = "…"

func filterTruncatecharsHelper(s string, newLen int) string {
	runes := []rune(s)
	if newLen < len(runes) {
		if newLen >= 1 {
			// Use proper ellipsis character (…) like Django does
			return string(runes[:newLen-1]) + ellipsis
		}
		// Django returns just the ellipsis for length <= 0
		return ellipsis
	}
	return string(runes)
}

// countHTMLTextRunes counts the number of text runes (non-tag characters)
// in an HTML string. This is used to determine whether truncation is needed.
func countHTMLTextRunes(value string) int {
	count := 0
	inTag := false
	for _, c := range value {
		if c == '<' {
			inTag = true
			continue
		}
		if c == '>' {
			inTag = false
			continue
		}
		if !inTag {
			count++
		}
	}
	return count
}

func filterTruncateHTMLHelper(value string, newOutput *bytes.Buffer, cond func() bool, fn func(c rune, s int, idx int) int, finalize func()) {
	vLen := len(value)
	var tagStack []string
	idx := 0

	for idx < vLen && !cond() {
		c, s := utf8.DecodeRuneInString(value[idx:])
		if c == utf8.RuneError {
			idx += s
			continue
		}

		if c == '<' {
			newOutput.WriteRune(c)
			idx += s // consume "<"

			if idx+1 < vLen {
				if value[idx] == '/' {
					// Close tag

					newOutput.WriteString("/")

					tag := ""
					idx++ // consume "/"

					for idx < vLen {
						c2, size2 := utf8.DecodeRuneInString(value[idx:])
						if c2 == utf8.RuneError {
							idx += size2
							continue
						}

						// End of tag found
						if c2 == '>' {
							idx++ // consume ">"
							break
						}
						tag += string(c2)
						idx += size2
					}

					if len(tagStack) > 0 {
						// Ideally, the close tag is TOP of tag stack
						// In malformed HTML, it must not be, so iterate through the stack and remove the tag
						for i := len(tagStack) - 1; i >= 0; i-- {
							if tagStack[i] == tag {
								// Found the tag
								tagStack[i] = tagStack[len(tagStack)-1]
								tagStack = tagStack[:len(tagStack)-1]
								break
							}
						}
					}

					newOutput.WriteString(tag)
					newOutput.WriteString(">")
				} else {
					// Open tag

					var tag strings.Builder

					params := false
					for idx < vLen {
						c2, size2 := utf8.DecodeRuneInString(value[idx:])
						if c2 == utf8.RuneError {
							idx += size2
							continue
						}

						newOutput.WriteRune(c2)

						// End of tag found
						if c2 == '>' {
							idx++ // consume ">"
							break
						}

						if !params {
							if c2 == ' ' {
								params = true
							} else {
								tag.WriteString(string(c2))
							}
						}

						idx += size2
					}

					// Add tag to stack
					tagStack = append(tagStack, tag.String())
				}
			}
		} else {
			idx = fn(c, s, idx)
		}
	}

	finalize()

	for i := len(tagStack) - 1; i >= 0; i-- {
		tag := tagStack[i]
		// Close everything from the regular tag stack
		fmt.Fprintf(newOutput, "</%s>", tag)
	}
}

// filterTruncatechars truncates a string if it is longer than the specified number
// of characters. Truncated strings will end with a translatable ellipsis character ("…").
// The ellipsis counts towards the character limit.
//
// Usage:
//
//	{{ "Joel is a slug"|truncatechars:7 }}
//
// Output: "Joel i…"
//
//	{{ "Hi"|truncatechars:5 }}
//
// Output: "Hi" (no truncation needed)
func filterTruncatechars(in *Value, param *Value) (*Value, error) {
	s := in.String()
	newLen := param.Integer()
	return AsValue(filterTruncatecharsHelper(s, newLen)), nil
}

// filterTruncatecharsHTML truncates a string if it is longer than the specified number
// of characters, similar to truncatechars but aware of HTML tags. Any tags that are
// opened in the string and not closed before the truncation point are closed immediately
// after the truncation. HTML tags are not counted towards the character limit.
// Truncated strings will end with an ellipsis character ("…") which counts towards the limit.
// Newlines in the HTML content will be preserved.
//
// Verified against Django 4.2 with script.
// Django difference: Returns AsSafeValue() to prevent double-escaping of preserved
// HTML tags. Django uses is_safe=True (preserves input safety status) instead.
//
// Usage:
//
//	{{ "<p>Joel is a slug</p>"|truncatechars_html:7 }}
//
// Output: "<p>Joel i…</p>"
func filterTruncatecharsHTML(in *Value, param *Value) (*Value, error) {
	value := in.String()
	maxLen := param.Integer()

	// Count the total number of text runes (excluding HTML tags) to determine
	// whether truncation is actually needed. Without this, we would always
	// reserve space for the ellipsis and truncate even when the full text
	// fits within the limit.
	totalTextRunes := countHTMLTextRunes(value)
	if totalTextRunes <= maxLen {
		// No truncation needed - return original value with tags intact
		return AsSafeValue(value), nil
	}

	// Reserve one character position for the ellipsis
	newLen := max(maxLen-1, 0)

	var newOutput bytes.Buffer

	textcounter := 0

	filterTruncateHTMLHelper(value, &newOutput, func() bool {
		return textcounter >= newLen
	}, func(c rune, s int, idx int) int {
		textcounter++
		newOutput.WriteRune(c)

		return idx + s
	}, func() {
		if textcounter >= newLen {
			newOutput.WriteString(ellipsis)
		}
	})

	return AsSafeValue(newOutput.String()), nil
}

// filterTruncatewords truncates a string after a certain number of words.
// If truncated, a space and Unicode ellipsis (" …") is appended.
//
// Django reference: django/utils/text.py Truncator.words(truncate=" …")
// Verified against Django 4.2 with script — space before ellipsis is intentional.
//
// Usage:
//
//	{{ "Hello beautiful world"|truncatewords:2 }}
//
// Output: "Hello beautiful …"
//
// {{ "Hi"|truncatewords:5 }}
//
// Output: "Hi"
func filterTruncatewords(in *Value, param *Value) (*Value, error) {
	words := strings.Fields(in.String())
	n := param.Integer()
	if n <= 0 {
		return AsValue(""), nil
	}
	nlen := min(len(words), n)
	out := make([]string, 0, nlen)
	for i := range nlen {
		out = append(out, words[i])
	}

	if n < len(words) {
		out = append(out, "\u2026")
	}

	return AsValue(strings.Join(out, " ")), nil
}

// filterTruncatewordsHTML truncates a string after a certain number of words,
// preserving HTML tags. HTML tags are not counted towards the word limit.
// If truncated, an ellipsis ("...") is appended. Open HTML tags are properly closed.
//
// Verified against Django 4.2 with script.
// Django difference: Returns AsSafeValue() to prevent double-escaping of preserved
// HTML tags. Django uses is_safe=True (preserves input safety status) instead.
//
// Usage:
//
//	{{ "<p>Hello beautiful world</p>"|truncatewords_html:2 }}
//
// Output: "<p>Hello beautiful …</p>"
func filterTruncatewordsHTML(in *Value, param *Value) (*Value, error) {
	value := in.String()
	newLen := max(param.Integer(), 0)

	newOutput := bytes.NewBuffer(nil)

	wordcounter := 0

	filterTruncateHTMLHelper(value, newOutput, func() bool {
		return wordcounter >= newLen
	}, func(_ rune, _ int, idx int) int {
		// Get next word
		wordFound := false

		for idx < len(value) {
			c2, size2 := utf8.DecodeRuneInString(value[idx:])
			if c2 == utf8.RuneError {
				idx += size2
				continue
			}

			if c2 == '<' {
				// HTML tag start, don't consume it
				return idx
			}

			newOutput.WriteRune(c2)
			idx += size2

			if c2 == ' ' || c2 == '.' || c2 == ',' || c2 == ';' {
				// Word ends here, stop capturing it now
				break
			} else {
				wordFound = true
			}
		}

		if wordFound {
			wordcounter++
		}

		return idx
	}, func() {
		if wordcounter >= newLen {
			newOutput.WriteString("\u2026")
		}
	})

	return AsSafeValue(newOutput.String()), nil
}

// filterEscape escapes a string's HTML characters. Specifically, it makes these replacements:
//   - < is converted to &lt;
//   - > is converted to &gt;
//   - ' (single quote) is converted to &#39;
//   - " (double quote) is converted to &quot;
//   - & is converted to &amp;
//
// The filter is also available under the alias "e".
//
// Usage:
//
//	{{ "<script>alert('XSS')</script>"|escape }}
//
// Output: "&lt;script&gt;alert(&#39;XSS&#39;)&lt;/script&gt;"
func filterEscape(in *Value, param *Value) (*Value, error) {
	return AsSafeValue(htmlEscapeReplacer.Replace(in.String())), nil
}

// filterSafe marks a string as safe, meaning it will not be HTML-escaped when
// rendered. Use this filter when you know the content is safe and should be
// rendered as-is (e.g., pre-sanitized HTML content).
//
// Usage:
//
//	{{ "<b>Bold text</b>"|safe }}
//
// Output: "<b>Bold text</b>"
//
// Without safe filter (when autoescape is on):
//
//	{{ "<b>Bold text</b>" }}
//
// Output: "&lt;b&gt;Bold text&lt;/b&gt;"
func filterSafe(in *Value, param *Value) (*Value, error) {
	return in, nil // nothing to do here, just to keep track of the safe application
}

// filterEscapejs escapes characters for safe use in JavaScript string literals.
// It converts special characters to their Unicode escape sequences (\uXXXX format).
//
// Characters that are escaped (matching Django's behavior):
//   - Backslash, quotes: \ ' " `
//   - HTML special chars: < > & = -
//   - Semicolon: ;
//   - Control characters: 0x00-0x1F, 0x7F, 0x80-0x9F
//   - Line separators: U+2028, U+2029
//
// Additionally, pongo2 interprets \r and \n escape sequences in the input
// and converts them to their Unicode escapes (\u000D and \u000A).
//
// Note: This filter escapes backticks, making it safe for JavaScript template
// literals as well as single/double quoted strings.
//
// Django reference: django/utils/html.py _js_escapes table.
// Verified against Django 4.2 with script.
// Django difference: uses lowercase hex (\u000d) vs Django's uppercase (\u000D) —
// functionally equivalent per JavaScript spec.
//
// Usage:
//
//	<script>var name = "{{ name|escapejs }}";</script>
//
// With name = "John's \"Quote\"":
//
// Output: <script>var name = "John\u0027s \u0022Quote\u0022";</script>
func filterEscapejs(in *Value, param *Value) (*Value, error) {
	sin := in.String()

	var b bytes.Buffer

	// Use index-based iteration to handle pongo2-specific \r and \n escape sequences
	idx := 0
	for idx < len(sin) {
		c, size := utf8.DecodeRuneInString(sin[idx:])
		if c == utf8.RuneError && size == 1 {
			// Invalid UTF-8, skip
			idx += size
			continue
		}

		// Handle pongo2-specific escape sequences: \r -> \u000D, \n -> \u000A
		if c == '\\' && idx+size < len(sin) {
			nextByte := sin[idx+size]
			switch nextByte {
			case 'r':
				b.WriteString(`\u000D`)
				idx += size + 1
				continue
			case 'n':
				b.WriteString(`\u000A`)
				idx += size + 1
				continue
			}
		}

		switch {
		// Characters that must be escaped for JavaScript string safety
		case c == '\\':
			b.WriteString(`\u005C`)
		case c == '\'':
			b.WriteString(`\u0027`)
		case c == '"':
			b.WriteString(`\u0022`)
		case c == '`':
			b.WriteString(`\u0060`)
		case c == '<':
			b.WriteString(`\u003C`)
		case c == '>':
			b.WriteString(`\u003E`)
		case c == '&':
			b.WriteString(`\u0026`)
		case c == '=':
			b.WriteString(`\u003D`)
		case c == '-':
			b.WriteString(`\u002D`)
		case c == ';':
			b.WriteString(`\u003B`)
		case c == '\u2028': // Line separator
			b.WriteString(`\u2028`)
		case c == '\u2029': // Paragraph separator
			b.WriteString(`\u2029`)
		// Control characters (0x00-0x1F, 0x7F, 0x80-0x9F)
		case c <= 0x1F, c == 0x7F, (c >= 0x80 && c <= 0x9F):
			fmt.Fprintf(&b, `\u%04X`, c)
		default:
			b.WriteRune(c)
		}

		idx += size
	}

	return AsSafeValue(b.String()), nil
}

// filterAdd adds the argument to the value. Works with numbers (integers and floats)
// and strings (concatenation).
//
// Usage with numbers:
//
//	{{ 5|add:3 }}
//
// Output: 8
//
//	{{ 3.5|add:2.1 }}
//
// Output: 5.600000
//
// Usage with strings:
//
//	{{ "Hello "|add:"World" }}
//
// Output: "Hello World"
func filterAdd(in *Value, param *Value) (*Value, error) {
	if in.IsNumber() && param.IsNumber() {
		if in.IsFloat() || param.IsFloat() {
			return AsValue(in.Float() + param.Float()), nil
		}
		return AsValue(in.Integer() + param.Integer()), nil
	}
	// If in/param is not a number, we're relying on the
	// Value's String() conversion and just add them both together
	return AsValue(in.String() + param.String()), nil
}

// filterAddslashes adds backslashes before quotes and backslashes.
// Useful for escaping strings in CSV or JavaScript contexts.
//
// Usage:
//
//	{{ "I'm using \"pongo2\""|addslashes }}
//
// Output: "I\'m using \"pongo2\""
func filterAddslashes(in *Value, param *Value) (*Value, error) {
	return AsValue(addslashesReplacer.Replace(in.String())), nil
}

// filterCut removes all occurrences of the argument from the string.
//
// Usage:
//
//	{{ "Hello World"|cut:" " }}
//
// Output: "HelloWorld"
//
//	{{ "String with spaces"|cut:" " }}
//
// Output: "Stringwithspaces"
func filterCut(in *Value, param *Value) (*Value, error) {
	return AsValue(strings.ReplaceAll(in.String(), param.String(), "")), nil
}

// filterLength returns the length of the value. Works with strings (character count),
// slices, arrays, and maps.
//
// Usage with strings:
//
//	{{ "Hello"|length }}
//
// Output: 5
//
// Usage with lists:
//
//	{% set items = ["a", "b", "c"] %}{{ items|length }}
//
// Output: 3
func filterLength(in *Value, param *Value) (*Value, error) {
	return AsValue(in.Len()), nil
}

// filterLengthis returns true if the value's length equals the argument.
// Useful in conditional expressions.
//
// Usage:
//
//	{% if items|length_is:3 %}Exactly 3 items{% endif %}
//
//	{{ "Hello"|length_is:5 }}
//
// Output: True
func filterLengthis(in *Value, param *Value) (*Value, error) {
	return AsValue(in.Len() == param.Integer()), nil
}

// filterDefault returns the argument if the value is falsy (empty string, 0,
// nil, false, empty slice/map). Otherwise returns the original value.
//
// Usage:
//
//	{{ name|default:"Guest" }}
//
// If name is empty or not set, output: "Guest"
// If name is "John", output: "John"
//
//	{{ 0|default:42 }}
//
// Output: 42
func filterDefault(in *Value, param *Value) (*Value, error) {
	if !in.IsTrue() {
		return param, nil
	}
	return in, nil
}

// filterDefaultIfNone returns the argument only if the value is nil.
// Unlike "default", this only triggers on nil values, not on other falsy values
// like 0, false, or empty strings.
//
// Usage:
//
//	{{ value|default_if_none:"N/A" }}
//
// If value is nil, output: "N/A"
// If value is 0, output: 0 (unlike default filter)
// If value is "", output: "" (unlike default filter)
func filterDefaultIfNone(in *Value, param *Value) (*Value, error) {
	if in.IsNil() {
		return param, nil
	}
	return in, nil
}

// filterDivisibleby returns true if the value is divisible by the argument.
// Returns false if the argument is 0 (to avoid division by zero).
//
// Usage:
//
//	{{ 21|divisibleby:7 }}
//
// Output: True
//
//	{% if forloop.Counter|divisibleby:2 %}even{% else %}odd{% endif %}
func filterDivisibleby(in *Value, param *Value) (*Value, error) {
	if param.Integer() == 0 {
		return AsValue(false), nil
	}
	return AsValue(in.Integer()%param.Integer() == 0), nil
}

// filterFirst returns the first element of a slice/array or the first character
// of a string. Returns an empty string if the input is empty.
//
// Usage with list:
//
//	{{ ["a", "b", "c"]|first }}
//
// Output: "a"
//
// Usage with string:
//
//	{{ "Hello"|first }}
//
// Output: "H"
func filterFirst(in *Value, param *Value) (*Value, error) {
	if in.CanSlice() && in.Len() > 0 {
		return in.Index(0), nil
	}
	return AsValue(""), nil
}

const maxFloatFormatDecimals = 1000

// filterFloatformat formats a floating-point number with a specified number of
// decimal places. If the argument is negative or omitted, trailing zeros are removed.
//
// Usage:
//
//	{{ 3.14159|floatformat:2 }}
//
// Output: "3.14"
//
//	{{ 3.0|floatformat:2 }}
//
// Output: "3.00"
//
//	{{ 3.0|floatformat:-2 }}
//
// Output: "3" (trailing zeros removed)
//
//	{{ 3.14159|floatformat }}
//
// Output: "3.1" (default: -1 decimal, trailing zeros removed)
func filterFloatformat(in *Value, param *Value) (*Value, error) {
	val := in.Float()

	decimals := -1
	if !param.IsNil() {
		// Any argument provided?
		decimals = param.Integer()
	}

	// if the argument is not a number (e. g. empty), the default
	// behaviour is trim the result
	trim := !param.IsNumber()

	if decimals <= 0 {
		// argument is negative or zero, so we
		// want the output being trimmed
		decimals = -decimals
		trim = true
	}

	if trim {
		// Remove zeroes
		if float64(int(val)) == val {
			return AsValue(in.Integer()), nil
		}
	}

	if decimals > maxFloatFormatDecimals {
		return nil, &Error{
			Sender:    "filter:floatformat",
			OrigError: fmt.Errorf("filter floatformat doesn't support more than %v decimals", maxFloatFormatDecimals),
		}
	}

	return AsValue(strconv.FormatFloat(val, 'f', decimals, 64)), nil
}

// filterGetdigit returns the digit at position N from the right (1-indexed).
// Position 1 is the rightmost digit. Returns the original value if N is out of range.
//
// Usage:
//
//	{{ 123456789|get_digit:1 }}
//
// Output: 9 (rightmost digit)
//
//	{{ 123456789|get_digit:2 }}
//
// Output: 8
//
//	{{ 123456789|get_digit:9 }}
//
// Output: 1 (leftmost digit)
func filterGetdigit(in *Value, param *Value) (*Value, error) {
	i := param.Integer()
	if i <= 0 {
		return in, nil
	}

	// Convert to string and validate it contains only digits (and optional leading minus).
	// This matches Django's behavior: int(value) must succeed, then we work with
	// the absolute value's digit string.
	s := in.String()

	// Determine the start of digits (skip optional leading minus sign)
	start := 0
	if len(s) > 0 && s[0] == '-' {
		start = 1
	}
	digits := s[start:]

	// Verify all remaining characters are digits; if not, return original value
	for j := 0; j < len(digits); j++ {
		if digits[j] < '0' || digits[j] > '9' {
			return in, nil
		}
	}

	l := len(digits)
	if l == 0 || i > l {
		return in, nil
	}

	return AsValue(int(digits[l-i] - '0')), nil
}

const filterIRIChars = "/#%[]=:;$&()+,!?*@'~"

// filterIriencode encodes an IRI (Internationalized Resource Identifier) for safe
// use in URLs. Unlike urlencode, it preserves characters that are valid in IRIs
// (such as /, #, %, etc.) while encoding other special characters using
// Go's url.QueryEscape.
//
// Django difference: Django's iri_to_uri() uses urllib.parse.quote() which encodes
// spaces as %20. Go's url.QueryEscape encodes spaces as +. Both are valid
// percent-encoding for query strings, but %20 is preferred in path segments.
//
// Usage:
//
//	{{ "https://example.com/path with spaces"|iriencode }}
//
// Output: "https://example.com/path+with+spaces"
//
//	{{ "/search?q=hello world"|iriencode }}
//
// Output: "/search?q=hello+world"
func filterIriencode(in *Value, param *Value) (*Value, error) {
	var b strings.Builder

	sin := in.String()
	for _, r := range sin {
		if strings.ContainsRune(filterIRIChars, r) {
			b.WriteRune(r)
		} else {
			b.WriteString(url.QueryEscape(string(r)))
		}
	}

	return AsValue(b.String()), nil
}

// filterJoin joins a list with the given separator string. For strings, each
// character is joined with the separator.
//
// Usage with list:
//
//	{{ ["apple", "banana", "cherry"]|join:", " }}
//
// Output: "apple, banana, cherry"
//
// Usage with string:
//
//	{{ "abc"|join:"-" }}
//
// Output: "a-b-c"
func filterJoin(in *Value, param *Value) (*Value, error) {
	if !in.CanSlice() {
		return in, nil
	}
	sep := param.String()
	if sep == "" {
		// An empty string separator returns the input string.
		return AsValue(in.String()), nil
	}

	sl := make([]string, 0, in.Len())

	// This is an optimization for very long strings. Index() splits `in` into runes with each
	// function invocation which hurts performance. Hence we're doing it just once (with ranging
	// over the string) and speeding things up.
	if in.IsString() {
		for _, i := range in.String() {
			sl = append(sl, string(i))
		}
	} else {
		for i := 0; i < in.Len(); i++ {
			sl = append(sl, in.Index(i).String())
		}
	}

	return AsValue(strings.Join(sl, sep)), nil
}

// filterLast returns the last element of a slice/array or the last character
// of a string. Returns an empty string if the input is empty.
//
// Usage with list:
//
//	{{ ["a", "b", "c"]|last }}
//
// Output: "c"
//
// Usage with string:
//
//	{{ "Hello"|last }}
//
// Output: "o"
func filterLast(in *Value, param *Value) (*Value, error) {
	if in.CanSlice() && in.Len() > 0 {
		return in.Index(in.Len() - 1), nil
	}
	return AsValue(""), nil
}

// filterUpper converts a string to uppercase.
//
// Usage:
//
//	{{ "Hello World"|upper }}
//
// Output: "HELLO WORLD"
func filterUpper(in *Value, param *Value) (*Value, error) {
	return AsValue(strings.ToUpper(in.String())), nil
}

// filterLower converts a string to lowercase.
//
// Usage:
//
//	{{ "Hello World"|lower }}
//
// Output: "hello world"
func filterLower(in *Value, param *Value) (*Value, error) {
	return AsValue(strings.ToLower(in.String())), nil
}

// filterMakelist converts a string into a list of individual characters.
// Each character becomes a separate element in the resulting list.
//
// Usage:
//
//	{{ "abc"|make_list }}
//
// Output: ["a", "b", "c"]
//
//	{% for char in "Hello"|make_list %}{{ char }}-{% endfor %}
//
// Output: "H-e-l-l-o-"
func filterMakelist(in *Value, param *Value) (*Value, error) {
	s := in.String()
	result := make([]string, 0, len(s))
	for _, c := range s {
		result = append(result, string(c))
	}
	return AsValue(result), nil
}

// filterCapfirst capitalizes the first character of a string.
// Only the first character is affected; the rest remains unchanged.
//
// Usage:
//
//	{{ "hello world"|capfirst }}
//
// Output: "Hello world"
//
//	{{ "hELLO"|capfirst }}
//
// Output: "HELLO"
func filterCapfirst(in *Value, param *Value) (*Value, error) {
	if in.Len() <= 0 {
		return AsValue(""), nil
	}
	t := in.String()
	r, size := utf8.DecodeRuneInString(t)
	return AsValue(strings.ToUpper(string(r)) + t[size:]), nil
}

const maxCharPadding = 10000

// filterCenter centers the value in a field of a given width by padding with spaces.
// If the original string is longer than the specified width, no padding is added.
//
// Usage:
//
//	"[{{ "hello"|center:11 }}]"
//
// Output: "[   hello   ]"
//
//	{{ "test"|center:10 }}
//
// Output: "   test   "
func filterCenter(in *Value, param *Value) (*Value, error) {
	width := param.Integer()
	slen := in.Len()
	if width <= slen {
		return in, nil
	}

	spaces := width - slen

	if spaces > maxCharPadding {
		return nil, &Error{
			Sender:    "filter:center",
			OrigError: fmt.Errorf("filter center doesn't support more than %v padding chars", maxCharPadding),
		}
	}

	// Match Python's str.center() padding bias:
	// When odd padding, extra space goes left if width is also odd, right otherwise.
	left := spaces/2 + (spaces & width & 1)
	right := spaces - left

	return AsValue(fmt.Sprintf("%s%s%s", strings.Repeat(" ", left),
		in.String(), strings.Repeat(" ", right))), nil
}

// filterDate formats a time.Time value according to the given Go time format string.
// This filter is also used for the "time" filter (same implementation).
// The format string uses Go's time formatting reference: Mon Jan 2 15:04:05 MST 2006.
//
// Django difference: Django uses PHP-style format characters (e.g., "Y-m-d" for
// "2024-03-15"); pongo2 uses Go's reference time format instead.
//
// Usage:
//
//	{{ myDate|date:"2006-01-02" }}
//
// Output: "2024-03-15" (example)
//
//	{{ myDate|date:"Monday, January 2, 2006" }}
//
// Output: "Friday, March 15, 2024" (example)
//
//	{{ myTime|time:"15:04:05" }}
//
// Output: "14:30:00" (example)
func filterDate(in *Value, param *Value) (*Value, error) {
	t, isTime := in.Interface().(time.Time)
	if !isTime {
		return nil, &Error{
			Sender:    "filter:date",
			OrigError: errors.New("filter input argument must be of type 'time.Time'"),
		}
	}
	return AsValue(t.Format(param.String())), nil
}

// filterFloat converts a value to a floating-point number.
// This is a pongo2-specific filter (not in Django).
//
// Usage:
//
//	{{ "3.14"|float }}
//
// Output: 3.140000
//
//	{{ 42|float }}
//
// Output: 42.000000
func filterFloat(in *Value, param *Value) (*Value, error) {
	return AsValue(in.Float()), nil
}

// filterInteger converts a value to an integer.
// This is a pongo2-specific filter (not in Django).
// Floating-point values are truncated.
//
// Usage:
//
//	{{ "42"|integer }}
//
// Output: 42
//
//	{{ 3.7|integer }}
//
// Output: 3
func filterInteger(in *Value, param *Value) (*Value, error) {
	return AsValue(in.Integer()), nil
}

// filterLinebreaks converts newlines in plain text to appropriate HTML.
// Single newlines become <br /> tags, and double newlines (blank lines)
// start a new paragraph with <p>...</p> tags.
//
// Uses the same algorithm as Django: split text on two or more consecutive
// newlines to form paragraphs, then replace remaining single newlines
// with <br /> within each paragraph.
//
// Usage:
//
//	{{ "First line\nSecond line"|linebreaks }}
//
// Output: "<p>First line<br />Second line</p>"
//
//	{{ "Para 1\n\nPara 2"|linebreaks }}
//
// Output: "<p>Para 1</p>\n\n<p>Para 2</p>"
func filterLinebreaks(in *Value, param *Value) (*Value, error) {
	s := normalizeNewlines(in.String())

	// Split on two or more consecutive newlines (paragraph breaks)
	paras := reDoubleNewline.Split(s, -1)

	var parts []string
	for _, para := range paras {
		para = strings.ReplaceAll(para, "\n", "<br />")
		parts = append(parts, "<p>"+para+"</p>")
	}

	return AsSafeValue(strings.Join(parts, "\n\n")), nil
}

// filterSplit splits a string by the given separator and returns a list.
//
// Usage:
//
//	{{ "a,b,c"|split:"," }}
//
// Output: ["a", "b", "c"]
//
//	{% for item in "one-two-three"|split:"-" %}{{ item }} {% endfor %}
//
// Output: "one two three "
func filterSplit(in *Value, param *Value) (*Value, error) {
	chunks := strings.Split(in.String(), param.String())

	return AsValue(chunks), nil
}

// filterLinebreaksbr converts all newlines in a string to HTML <br /> tags.
// Unlike linebreaks, this filter does not wrap text in <p> tags.
//
// Usage:
//
//	{{ "First line\nSecond line\nThird line"|linebreaksbr }}
//
// Output: "First line<br />Second line<br />Third line"
func filterLinebreaksbr(in *Value, param *Value) (*Value, error) {
	s := normalizeNewlines(in.String())
	return AsSafeValue(strings.ReplaceAll(s, "\n", "<br />")), nil
}

// filterLinenumbers prepends line numbers to each line in the text.
// Line numbering starts at 1.
//
// Usage:
//
//	{{ "first\nsecond\nthird"|linenumbers }}
//
// Output:
//
//  1. first
//  2. second
//  3. third
func filterLinenumbers(in *Value, param *Value) (*Value, error) {
	s := normalizeNewlines(in.String())
	lines := strings.Split(s, "\n")
	// Calculate padding width for zero-padded line numbers (matching Django)
	width := len(strconv.Itoa(len(lines)))
	output := make([]string, 0, len(lines))
	for idx, line := range lines {
		output = append(output, fmt.Sprintf("%0*d. %s", width, idx+1, line))
	}
	return AsValue(strings.Join(output, "\n")), nil
}

// filterLjust left-aligns the value in a field of a given width by padding
// spaces on the right. If the original string is longer than the specified width,
// no padding is added.
//
// Usage:
//
//	"[{{ "hello"|ljust:10 }}]"
//
// Output: "[hello     ]"
func filterLjust(in *Value, param *Value) (*Value, error) {
	times := max(param.Integer()-in.Len(), 0)
	if times > maxCharPadding {
		return nil, &Error{
			Sender:    "filter:ljust",
			OrigError: fmt.Errorf("ljust doesn't support more padding than %d chars", maxCharPadding),
		}
	}
	return AsValue(fmt.Sprintf("%s%s", in.String(), strings.Repeat(" ", times))), nil
}

// filterUrlencode encodes a string for safe use in a URL query string
// using Go's url.QueryEscape.
//
// Django difference: Django's urlencode uses urllib.parse.quote(safe='/') which
// encodes spaces as %20 and preserves forward slashes. Go's url.QueryEscape
// encodes spaces as + and encodes forward slashes as %2F.
//
// Usage:
//
//	{{ "hello world"|urlencode }}
//
// Output: "hello+world"
//
//	{{ "http://example.org/path?a=b"|urlencode }}
//
// Output: "http%3A%2F%2Fexample.org%2Fpath%3Fa%3Db"
func filterUrlencode(in *Value, param *Value) (*Value, error) {
	return AsValue(url.QueryEscape(in.String())), nil
}

// normalizeNewlines converts \r\n and lone \r to \n.
func normalizeNewlines(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return strings.ReplaceAll(s, "\r", "\n")
}

var (
	// URL regex matches:
	// 1. URLs starting with http:// or https://
	// 2. URLs starting with www.
	// 3. Bare domains with common TLDs (generic, country-code, and new TLDs)
	filterUrlizeURLRegexp = regexp.MustCompile(`((((http|https)://)|www\.|((^|[ ])[0-9A-Za-z_\-]+\.(com|net|org|info|biz|edu|gov|mil|int|co|io|ai|app|dev|me|tv|cc|us|uk|de|fr|es|it|nl|be|at|ch|ru|cn|jp|kr|au|nz|in|br|mx|ca|eu))))\S*([ ]+|$)`)
	// Email regex matches email addresses with TLDs 2-6 chars to support .info, .museum, etc.
	filterUrlizeEmailRegexp = regexp.MustCompile(`(\w+@\w+\.\w{2,6})`)
)

func filterUrlizeHelper(input string, autoescape bool, trunc int) (string, error) {
	var soutErr error
	sout := filterUrlizeURLRegexp.ReplaceAllStringFunc(input, func(raw_url string) string {
		var prefix string
		var suffix string
		if strings.HasPrefix(raw_url, " ") {
			prefix = " "
		}
		if strings.HasSuffix(raw_url, " ") {
			suffix = " "
		}

		raw_url = strings.TrimSpace(raw_url)

		t, err := ApplyFilter("iriencode", AsValue(raw_url), nil)
		if err != nil {
			soutErr = err
			return ""
		}
		url := t.String()

		if !strings.HasPrefix(url, "http") {
			url = fmt.Sprintf("http://%s", url)
		}

		title := raw_url

		titleRunes := []rune(title)
		if trunc > 1 && len(titleRunes) > trunc {
			title = string(titleRunes[:trunc-1]) + ellipsis
		}

		if autoescape {
			t, err := ApplyFilter("escape", AsValue(title), nil)
			if err != nil {
				soutErr = err
				return ""
			}
			title = t.String()
		}

		return fmt.Sprintf(`%s<a href="%s" rel="nofollow">%s</a>%s`, prefix, url, title, suffix)
	})
	if soutErr != nil {
		return "", soutErr
	}

	sout = filterUrlizeEmailRegexp.ReplaceAllStringFunc(sout, func(mail string) string {
		title := mail

		titleRunes := []rune(title)
		if trunc > 1 && len(titleRunes) > trunc {
			title = string(titleRunes[:trunc-1]) + ellipsis
		}

		return fmt.Sprintf(`<a href="mailto:%s">%s</a>`, mail, title)
	})

	return sout, nil
}

// filterUrlize converts URLs and email addresses in plain text into clickable links.
// URLs are wrapped in <a> tags with rel="nofollow". Email addresses become mailto: links.
// By default, the links are HTML-escaped; pass false to disable escaping.
//
// Usage:
//
//	{{ "Visit www.example.com today!"|urlize }}
//
// Output: 'Visit <a href="http://www.example.com" rel="nofollow">www.example.com</a> today!'
//
//	{{ "Contact: user@example.com"|urlize }}
//
// Output: 'Contact: <a href="mailto:user@example.com">user@example.com</a>'
func filterUrlize(in *Value, param *Value) (*Value, error) {
	autoescape := true
	if param.IsBool() {
		autoescape = param.Bool()
	}

	s, err := filterUrlizeHelper(in.String(), autoescape, -1)
	if err != nil {
		return nil, &Error{
			Sender:    "filter:urlize",
			OrigError: err,
		}
	}

	return AsSafeValue(s), nil
}

// filterUrlizetrunc works like urlize but truncates URLs longer than the given
// character limit. An ellipsis is appended to truncated URLs.
//
// Usage:
//
//	{{ "Check out www.reallylongdomainname.com/path"|urlizetrunc:20 }}
//
// Output: 'Check out <a href="http://www.reallylongdomainname.com/path" rel="nofollow">www.reallylongdomai…</a>'
func filterUrlizetrunc(in *Value, param *Value) (*Value, error) {
	s, err := filterUrlizeHelper(in.String(), true, param.Integer())
	if err != nil {
		return nil, &Error{
			Sender:    "filter:urlizetrunc",
			OrigError: err,
		}
	}
	return AsSafeValue(s), nil
}

// filterStringformat formats the value according to the argument, which is a
// Go fmt-style format specifier.
//
// Django difference: Django uses Python % formatting (e.g., "%.2f"); pongo2 uses
// Go fmt verbs (e.g., "%.2f" works the same, but "%s" vs "%s" differ in edge cases).
//
// Usage:
//
//	{{ 3.14159|stringformat:"%.2f" }}
//
// Output: "3.14"
//
//	{{ 42|stringformat:"%05d" }}
//
// Output: "00042"
//
//	{{ "hello"|stringformat:"%q" }}
//
// Output: '"hello"'
func filterStringformat(in *Value, param *Value) (*Value, error) {
	return AsValue(fmt.Sprintf(param.String(), in.Interface())), nil
}

// reStriptags matches HTML/XML tags including those with quoted attributes containing >.
// Pattern breakdown:
// - < : opening angle bracket
// - [a-zA-Z!/?\[] : tag must start with letter, !, /, ?, or [ (for CDATA/comments)
// - (?: ... )* : non-capturing group for tag content, zero or more times
//   - "[^"]*" : double-quoted string (can contain >)
//   - '[^']*' : single-quoted string (can contain >)
//   - [^>] : any char except >
//
// - > : closing angle bracket
var reDoubleNewline = regexp.MustCompile(`\n{2,}`)
var reStriptags = regexp.MustCompile(`<[a-zA-Z!/?\[](?:"[^"]*"|'[^']*'|[^>])*>`)

// filterStriptags strips all HTML/XML tags from the value, returning plain text.
// Null bytes are removed from the input, and the result is trimmed of leading/trailing whitespace.
//
// SECURITY WARNING: This filter does NOT guarantee HTML-safe output, particularly
// with malformed or malicious HTML input. Never apply the |safe filter to striptags
// output. For security-critical applications, use a proper HTML sanitization library.
//
// Usage:
//
//	{{ "<p>Hello <b>World</b>!</p>"|striptags }}
//
// Output: "Hello World!"
//
//	{{ "<a href='#'>Link</a>"|striptags }}
//
// Output: "Link"
func filterStriptags(in *Value, param *Value) (*Value, error) {
	s := in.String()

	// Remove null bytes which could be used to bypass filters
	s = strings.ReplaceAll(s, "\x00", "")

	result, err := stripTagsIteratively(s, []*regexp.Regexp{reStriptags}, 50, "filter:striptags")
	if err != nil {
		return nil, err
	}

	return AsValue(result), nil
}

// https://en.wikipedia.org/wiki/Phoneword
var filterPhone2numericMap = map[string]string{
	"a": "2", "b": "2", "c": "2", "d": "3", "e": "3", "f": "3", "g": "4", "h": "4", "i": "4", "j": "5", "k": "5",
	"l": "5", "m": "6", "n": "6", "o": "6", "p": "7", "q": "7", "r": "7", "s": "7", "t": "8", "u": "8", "v": "8",
	"w": "9", "x": "9", "y": "9", "z": "9",
}

// filterPhone2numeric converts a phone number with letters (phoneword) to its
// numeric equivalent using the standard phone keypad mapping.
// See: https://en.wikipedia.org/wiki/Phoneword
//
// Usage:
//
//	{{ "1-800-COLLECT"|phone2numeric }}
//
// Output: "1-800-2655328"
//
//	{{ "CALL-ME"|phone2numeric }}
//
// Output: "2255-63"
func filterPhone2numeric(in *Value, param *Value) (*Value, error) {
	sin := in.String()
	for k, v := range filterPhone2numericMap {
		sin = strings.ReplaceAll(sin, k, v)
		sin = strings.ReplaceAll(sin, strings.ToUpper(k), v)
	}
	return AsValue(sin), nil
}

// filterPluralize returns a plural suffix based on the numeric value.
// By default, returns "s" if the value is not 1, otherwise returns "".
// You can specify custom singular/plural suffixes as comma-separated arguments.
//
// Usage:
//
//	You have {{ count }} item{{ count|pluralize }}.
//
// With count=1: "You have 1 item."
// With count=5: "You have 5 items."
//
//	{{ count }} cherr{{ count|pluralize:"y,ies" }}.
//
// With count=1: "1 cherry."
// With count=5: "5 cherries."
//
//	{{ count }} walrus{{ count|pluralize:"es" }}.
//
// With count=1: "1 walrus."
// With count=5: "5 walruses."
func filterPluralize(in *Value, param *Value) (*Value, error) {
	if in.IsNumber() {
		// Use Float() comparison instead of Integer() to avoid truncating
		// floats like 1.5 to 1, which would incorrectly treat them as singular.
		isPlural := in.Float() != 1

		// Works only on numbers
		if param.Len() > 0 {
			endings := strings.Split(param.String(), ",")
			if len(endings) > 2 {
				return nil, &Error{
					Sender:    "filter:pluralize",
					OrigError: errors.New("you cannot pass more than 2 arguments to filter 'pluralize'"),
				}
			}
			if len(endings) == 1 {
				// 1 argument
				if isPlural {
					return AsValue(endings[0]), nil
				}
			} else {
				if isPlural {
					// 2 arguments
					return AsValue(endings[1]), nil
				}
				return AsValue(endings[0]), nil
			}
		} else {
			if isPlural {
				// return default 's'
				return AsValue("s"), nil
			}
		}

		return AsValue(""), nil
	}
	return nil, &Error{
		Sender:    "filter:pluralize",
		OrigError: errors.New("filter 'pluralize' does only work on numbers"),
	}
}

// filterRandom returns a random element from the given list or string.
// If the input is empty, returns the input unchanged.
//
// Usage:
//
//	{{ ["apple", "banana", "cherry"]|random }}
//
// Output: "banana" (random element)
//
//	{{ "abc"|random }}
//
// Output: "b" (random character)
func filterRandom(in *Value, param *Value) (*Value, error) {
	if !in.CanSlice() || in.Len() <= 0 {
		return in, nil
	}
	i := rand.Intn(in.Len())
	return in.Index(i), nil
}

var reTagName = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9]*$`)

// filterRemovetags removes specified HTML tags from the string while keeping the content.
// Tag names are provided as a comma-separated list.
//
// SECURITY WARNING: While this implementation applies stripping recursively to handle
// obfuscated tags like "<sc<script>ript>", it is NOT guaranteed to be XSS-safe.
// For security-critical applications, use a proper HTML sanitization library instead.
//
// This filter was removed from Django 1.10. See:
// https://www.djangoproject.com/weblog/2014/aug/11/remove-tags-advisory/
//
// Usage:
//
//	{{ "<b>bold</b> and <i>italic</i>"|removetags:"b" }}
//
// Output: "bold and <i>italic</i>"
//
//	{{ "<script>alert('xss')</script>"|removetags:"script" }}
//
// Output: "alert('xss')"
//
// Note: For XSS prevention, use a proper HTML sanitization library.
func filterRemovetags(in *Value, param *Value) (*Value, error) {
	s := in.String()
	tags := strings.Split(param.String(), ",")

	// Build regex patterns for all specified tags
	var patterns []*regexp.Regexp
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		if !reTagName.MatchString(tag) {
			return nil, &Error{
				Sender:    "filter:removetags",
				OrigError: fmt.Errorf("invalid tag name '%s'", tag),
			}
		}

		// Match opening tags (with optional attributes), closing tags, and self-closing tags
		// Case-insensitive matching
		// Pattern matches: <tag>, <tag attr>, </tag>, <tag/>, <tag />
		re, err := regexp.Compile(fmt.Sprintf(`(?i)</?%s(?:\s[^>]*)?/?>`, regexp.QuoteMeta(tag)))
		if err != nil {
			return nil, &Error{
				Sender:    "filter:removetags",
				OrigError: fmt.Errorf("removetags-filter regexp error with tag '%s': %v", tag, err),
			}
		}
		patterns = append(patterns, re)
	}

	result, err := stripTagsIteratively(s, patterns, 100, "filter:removetags")
	if err != nil {
		return nil, err
	}

	return AsValue(result), nil
}

// filterRjust right-aligns the value in a field of a given width by padding
// spaces on the left. Useful for creating aligned columns of text.
//
// Usage:
//
//	"[{{ "hello"|rjust:10 }}]"
//
// Output: "[     hello]"
//
//	{{ 42|rjust:5 }}
//
// Output: "   42"
func filterRjust(in *Value, param *Value) (*Value, error) {
	padding := param.Integer()
	if padding > maxCharPadding {
		return nil, &Error{
			Sender:    "filter:rjust",
			OrigError: fmt.Errorf("rjust doesn't support more padding than %d chars", maxCharPadding),
		}
	}
	return AsValue(fmt.Sprintf(fmt.Sprintf("%%%ds", padding), in.String())), nil
}

// filterSlice returns a slice of a list using the "from:to" syntax (Python-style).
// Both from and to are optional. Negative indices count from the end.
//
// Usage:
//
//	{{ [1, 2, 3, 4, 5]|slice:"1:3" }}
//
// Output: [2, 3]
//
//	{{ [1, 2, 3, 4, 5]|slice:":3" }}
//
// Output: [1, 2, 3]
//
//	{{ [1, 2, 3, 4, 5]|slice:"2:" }}
//
// Output: [3, 4, 5]
//
//	{{ [1, 2, 3, 4, 5]|slice:"-2:" }}
//
// Output: [4, 5]
//
//	{{ "Hello"|slice:"1:4" }}
//
// Output: "ell"
func filterSlice(in *Value, param *Value) (*Value, error) {
	comp := strings.Split(param.String(), ":")
	if len(comp) != 2 {
		return nil, &Error{
			Sender:    "filter:slice",
			OrigError: errors.New("Slice string must have the format 'from:to' [from/to can be omitted, but the ':' is required]"),
		}
	}

	if !in.CanSlice() {
		return in, nil
	}

	// start with [x:len]
	from := AsValue(comp[0]).Integer()
	to := in.Len()

	// handle negative x
	if from < 0 {
		from = max(in.Len()+from, 0)
	}

	// handle x over bounds
	if from > to {
		from = to
	}

	vto := AsValue(comp[1]).Integer()
	// handle missing y
	if strings.TrimSpace(comp[1]) == "" {
		vto = in.Len()
	}

	// handle negative y
	if vto < 0 {
		vto = max(in.Len()+vto, 0)
	}

	// handle y < x
	if vto < from {
		vto = from
	}

	// y is within bounds, return the [x, y] slice
	if vto >= from && vto <= in.Len() {
		to = vto
	} // otherwise, the slice remains [x, len]

	return in.Slice(from, to), nil
}

// reTitleApostrophe matches a lowercase letter followed by an apostrophe and an uppercase letter.
// Used to fix Python's str.title() behavior with apostrophes, e.g., "It'S" -> "It's".
var reTitleApostrophe = regexp.MustCompile(`([a-z])'([A-Z])`)

// reTitleDigit matches a digit followed by an uppercase letter.
// Used to fix titlecase after digits, e.g., "1St" -> "1st".
var reTitleDigit = regexp.MustCompile(`(\d)([A-Z])`)

// filterTitle converts a string to title case, where the first character of
// each word is capitalized and the rest are lowercase. Matches Django's behavior:
// capitalizes after any non-alphanumeric character (including underscores and hyphens),
// but not after apostrophes within words or after digits.
//
// Usage:
//
//	{{ "hello world"|title }}
//
// Output: "Hello World"
//
//	{{ "HELLO WORLD"|title }}
//
// Output: "Hello World"
//
//	{{ "hello_world"|title }}
//
// Output: "Hello_World"
//
//	{{ "it's a test"|title }}
//
// Output: "It's A Test"
func filterTitle(in *Value, param *Value) (*Value, error) {
	if !in.IsString() {
		return AsValue(""), nil
	}
	s := in.String()

	// Titlecase: capitalize the first letter after any non-alphanumeric character.
	// This matches Python's str.title() behavior.
	runes := []rune(s)
	capitalizeNext := true
	for i, r := range runes {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			capitalizeNext = true
		} else if capitalizeNext {
			runes[i] = unicode.ToUpper(r)
			capitalizeNext = false
		} else {
			runes[i] = unicode.ToLower(r)
		}
	}
	result := string(runes)

	// Fix apostrophe case: "It'S" -> "It's" (Django regex: ([a-z])'([A-Z]))
	result = reTitleApostrophe.ReplaceAllStringFunc(result, func(m string) string {
		return strings.ToLower(m)
	})

	// Fix digit case: "1St" -> "1st" (Django regex: \d([A-Z]))
	result = reTitleDigit.ReplaceAllStringFunc(result, func(m string) string {
		return strings.ToLower(m)
	})

	return AsValue(result), nil
}

// filterWordcount returns the number of words in the string.
// Words are separated by whitespace.
//
// Usage:
//
//	{{ "Hello beautiful world"|wordcount }}
//
// Output: 3
//
//	{{ "  Multiple   spaces  "|wordcount }}
//
// Output: 2
func filterWordcount(in *Value, param *Value) (*Value, error) {
	return AsValue(len(strings.Fields(in.String()))), nil
}

// filterWordwrap wraps text at the specified character column width.
// Lines are broken at word boundaries (words are not split). Existing
// newlines are preserved. Long words that exceed the width are not broken.
// \r\n and \r are normalized to \n before processing.
//
// Verified against Django 4.2 django.utils.text.wrap().
// Django ref: django/utils/text.py wrap()
//
// Usage:
//
//	{{ "a b c d e f g h"|wordwrap:5 }}
//
// Output:
//
//	a b c
//	d e f
//	g h
func filterWordwrap(in *Value, param *Value) (*Value, error) {
	s := normalizeNewlines(in.String())
	wrapAt := param.Integer()
	if wrapAt <= 0 {
		return in, nil
	}

	// Preserve existing line breaks, wrap each line independently.
	// Long words are not broken (matching Django's break_long_words=False).
	inputLines := strings.Split(s, "\n")
	var resultLines []string

	for _, line := range inputLines {
		if line == "" {
			resultLines = append(resultLines, line)
			continue
		}
		words := strings.Fields(line)
		if len(words) == 0 {
			// Line contains only whitespace; preserve it
			resultLines = append(resultLines, line)
			continue
		}

		var currentLine strings.Builder
		currentLine.WriteString(words[0])
		currentLen := utf8.RuneCountInString(words[0])

		for _, word := range words[1:] {
			wordLen := utf8.RuneCountInString(word)
			// +1 for the space between words
			if currentLen+1+wordLen > wrapAt {
				resultLines = append(resultLines, currentLine.String())
				currentLine.Reset()
				currentLine.WriteString(word)
				currentLen = wordLen
			} else {
				currentLine.WriteString(" ")
				currentLine.WriteString(word)
				currentLen += 1 + wordLen
			}
		}
		resultLines = append(resultLines, currentLine.String())
	}

	return AsValue(strings.Join(resultLines, "\n")), nil
}

// filterYesno maps true, false, and nil values to customizable strings.
// By default: true -> "yes", false -> "no", nil -> "maybe".
// You can provide custom values as comma-separated arguments: "yes_val,no_val,maybe_val".
//
// Usage:
//
//	{{ true|yesno }}
//
// Output: "yes"
//
//	{{ false|yesno }}
//
// Output: "no"
//
//	{{ nil|yesno }}
//
// Output: "maybe"
//
//	{{ true|yesno:"yeah,nope,dunno" }}
//
// Output: "yeah"
//
//	{{ false|yesno:"on,off" }}
//
// Output: "off"
func filterYesno(in *Value, param *Value) (*Value, error) {
	choices := map[int]string{
		0: "yes",
		1: "no",
		2: "maybe",
	}
	paramString := param.String()
	customChoices := strings.Split(paramString, ",")
	if len(paramString) > 0 {
		if len(customChoices) > 3 {
			return nil, &Error{
				Sender:    "filter:yesno",
				OrigError: fmt.Errorf("you cannot pass more than 3 options to the 'yesno'-filter (got: '%s')", paramString),
			}
		}
		if len(customChoices) < 2 {
			return nil, &Error{
				Sender:    "filter:yesno",
				OrigError: fmt.Errorf("you must either pass no or at least 2 arguments to the 'yesno'-filter (got: '%s')", paramString),
			}
		}

		// Map to the options now
		choices[0] = customChoices[0]
		choices[1] = customChoices[1]
		if len(customChoices) == 3 {
			choices[2] = customChoices[2]
		} else {
			// Django: with only 2 args, nil maps to the "no" value (same as false)
			choices[2] = customChoices[1]
		}
	}

	// maybe
	if in.IsNil() {
		return AsValue(choices[2]), nil
	}

	// yes
	if in.IsTrue() {
		return AsValue(choices[0]), nil
	}

	// no
	return AsValue(choices[1]), nil
}

// timeFilterHelper extracts the common logic for timesince/timeuntil filters.
// When reverse is false, computes timesince (elapsed time from d to now).
// When reverse is true, computes timeuntil (remaining time from now to d).
func timeFilterHelper(in *Value, param *Value, reverse bool) (*Value, error) {
	t, isTime := in.Interface().(time.Time)
	if !isTime {
		return AsValue(""), nil
	}

	comparisonTime := time.Now()
	if !param.IsNil() {
		if paramTime, ok := param.Interface().(time.Time); ok {
			comparisonTime = paramTime
		}
	}

	var from, to time.Time
	if reverse {
		from, to = comparisonTime, t
	} else {
		from, to = t, comparisonTime
	}

	// If 'from' is after 'to', the direction is wrong: return "0 minutes".
	// Django: timesince returns "0 minutes" for future dates,
	// timeuntil returns "0 minutes" for past dates.
	if to.Before(from) {
		return AsValue("0 minutes"), nil
	}

	return AsValue(timeDiff(from, to)), nil
}

// filterTimesince returns the time elapsed since the given datetime.
// The result is a human-readable string like "2 days, 3 hours".
//
// Usage:
//
//	{{ some_date|timesince }}
//	{{ some_date|timesince:comparison_date }}
func filterTimesince(in *Value, param *Value) (*Value, error) {
	return timeFilterHelper(in, param, false)
}

// filterTimeuntil returns the time remaining until the given datetime.
// The result is a human-readable string like "2 days, 3 hours".
//
// Usage:
//
//	{{ some_date|timeuntil }}
//	{{ some_date|timeuntil:comparison_date }}
func filterTimeuntil(in *Value, param *Value) (*Value, error) {
	return timeFilterHelper(in, param, true)
}

// monthsDays maps month index (0-based) to number of days in that month (non-leap year).
var monthsDays = [12]int{31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}

// timeDiff calculates the difference between two times and returns a human-readable string.
// Uses the same algorithm as Django: calendar-based year/month calculation with a pivot
// date, and only shows up to two adjacent time units.
func timeDiff(from, to time.Time) string {
	since := to.Sub(from)
	if since < time.Minute {
		return "0 minutes"
	}

	// Calculate years and months using calendar arithmetic (like Django).
	totalMonths := (to.Year()-from.Year())*12 + int(to.Month()) - int(from.Month())
	if from.Day() > to.Day() || (from.Day() == to.Day() && timeOfDay(from) > timeOfDay(to)) {
		totalMonths--
	}
	if totalMonths < 0 {
		totalMonths = 0
	}
	years := totalMonths / 12
	months := totalMonths % 12

	// Create a pivot date shifted by years+months from 'from', then calculate
	// the remaining duration from pivot to 'to'.
	var remaining time.Duration
	if years > 0 || months > 0 {
		pivotYear := from.Year() + years
		pivotMonth := int(from.Month()) + months
		if pivotMonth > 12 {
			pivotMonth -= 12
			pivotYear++
		}
		maxDay := monthsDays[pivotMonth-1]
		if pivotMonth == 2 && isLeapYear(pivotYear) {
			maxDay = 29
		}
		pivotDay := from.Day()
		if pivotDay > maxDay {
			pivotDay = maxDay
		}
		pivot := time.Date(pivotYear, time.Month(pivotMonth), pivotDay,
			from.Hour(), from.Minute(), from.Second(), 0, from.Location())
		remaining = to.Sub(pivot)
		if remaining < 0 {
			remaining = 0
		}
	} else {
		remaining = since
	}

	weeks := int(remaining / (7 * 24 * time.Hour))
	remaining -= time.Duration(weeks) * 7 * 24 * time.Hour

	days := int(remaining / (24 * time.Hour))
	remaining -= time.Duration(days) * 24 * time.Hour

	hours := int(remaining / time.Hour)
	remaining -= time.Duration(hours) * time.Hour

	minutes := int(remaining / time.Minute)

	// Collect units in order: years, months, weeks, days, hours, minutes.
	// Django only shows up to 2 adjacent units (e.g., "1 year, 2 months"
	// but not "1 year, 3 days" since months would be skipped).
	type unit struct {
		value int
		name  string
	}
	units := []unit{
		{years, "year"},
		{months, "month"},
		{weeks, "week"},
		{days, "day"},
		{hours, "hour"},
		{minutes, "minute"},
	}

	var parts []string
	lastIdx := -2 // track index of last added unit for adjacency check
	for i, u := range units {
		if u.value <= 0 {
			continue
		}
		// Enforce adjacency: only add if this unit is adjacent to the last one added
		if len(parts) > 0 && i != lastIdx+1 {
			break
		}
		if len(parts) >= 2 {
			break
		}
		name := u.name
		if u.value != 1 {
			name += "s"
		}
		parts = append(parts, fmt.Sprintf("%d %s", u.value, name))
		lastIdx = i
	}

	if len(parts) == 0 {
		return "0 minutes"
	}

	return strings.Join(parts, ", ")
}

// timeOfDay returns the time-of-day portion as a duration for comparison.
func timeOfDay(t time.Time) time.Duration {
	return time.Duration(t.Hour())*time.Hour +
		time.Duration(t.Minute())*time.Minute +
		time.Duration(t.Second())*time.Second
}

// isLeapYear returns true if the given year is a leap year.
func isLeapYear(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

// filterDictsort sorts a list of maps or structs by the specified key.
//
// Usage:
//
//	{{ items|dictsort:"name" }}
//
// For a list of maps, this sorts by the value of the specified key.
// For a list of structs, this sorts by the specified field name.
func filterDictsort(in *Value, param *Value) (*Value, error) {
	return dictsortHelper(in, param, false)
}

// filterDictsortReversed sorts a list of maps or structs by the specified key in reverse order.
//
// Usage:
//
//	{{ items|dictsortreversed:"name" }}
func filterDictsortReversed(in *Value, param *Value) (*Value, error) {
	return dictsortHelper(in, param, true)
}

// dictsortItems implements sort.Interface for sorting by key.
// It uses type-aware comparison: numeric keys are compared numerically,
// all other keys are compared as strings.
type dictsortItems struct {
	entries []struct {
		item   *Value
		sortBy *Value
	}
	allNumeric bool
}

func (d dictsortItems) Len() int      { return len(d.entries) }
func (d dictsortItems) Swap(i, j int) { d.entries[i], d.entries[j] = d.entries[j], d.entries[i] }
func (d dictsortItems) Less(i, j int) bool {
	if d.allNumeric {
		return d.entries[i].sortBy.Float() < d.entries[j].sortBy.Float()
	}
	return d.entries[i].sortBy.String() < d.entries[j].sortBy.String()
}

func dictsortHelper(in *Value, param *Value, reverse bool) (*Value, error) {
	if !in.CanSlice() {
		return in, nil
	}

	if param.IsNil() {
		return nil, errors.New("dictsort requires a key argument")
	}

	// Collect items with their sort keys
	items := dictsortItems{allNumeric: true}

	in.Iterate(func(idx, count int, k, value *Value) bool {
		// Get the item (value for maps, key for slices/arrays)
		item := value
		if item == nil {
			item = k
		}

		// Get the sort key value using Value methods
		sortBy := AsValue("")
		if item.IsMap() || item.IsStruct() {
			sortVal := item.GetItem(param)
			if !sortVal.IsNil() {
				sortBy = sortVal
			}
		}

		if items.allNumeric && !sortBy.IsNumber() {
			items.allNumeric = false
		}

		items.entries = append(items.entries, struct {
			item   *Value
			sortBy *Value
		}{item: item, sortBy: sortBy})
		return true
	}, func() {})

	// Sort by the key
	if reverse {
		sort.Sort(sort.Reverse(items))
	} else {
		sort.Sort(items)
	}

	// Build result
	var result []any
	for _, entry := range items.entries {
		result = append(result, entry.item.Interface())
	}

	return AsValue(result), nil
}

// filterUnorderedList recursively generates an unordered HTML list from nested lists.
//
// Usage:
//
//	{{ items|unordered_list }}
//
// For input: ["States", ["Kansas", ["Lawrence", "Topeka"], "Illinois"]]
// Output: <li>States<ul><li>Kansas<ul><li>Lawrence</li><li>Topeka</li></ul></li><li>Illinois</li></ul></li>
//
// filterUnorderedList outputs the inner list items only (without wrapping <ul></ul> tags),
// with tab indentation matching Django's format. Each nesting level adds one tab.
//
// Django reference: django/template/defaultfilters.py list_formatter()
func filterUnorderedList(in *Value, param *Value) (*Value, error) {
	if !in.IsSliceOrArray() {
		return AsSafeValue(""), nil
	}
	result := unorderedListFormatter(in, 1)
	return AsSafeValue(result), nil
}

const maxUnorderedListDepth = 100

// unorderedListFormatter formats a nested list with tab indentation, matching Django's
// list_formatter function. tabs starts at 1 and increments for each nesting level.
func unorderedListFormatter(in *Value, tabs int) string {
	if tabs > maxUnorderedListDepth {
		return ""
	}

	if !in.IsSliceOrArray() {
		return ""
	}

	// Collect all items from the list
	items := make([]*Value, 0)
	in.Iterate(func(idx, count int, key, value *Value) bool {
		if value != nil {
			items = append(items, value)
		} else {
			items = append(items, key)
		}
		return true
	}, func() {})

	indent := strings.Repeat("\t", tabs)
	var output []string

	// Walk items, pairing each non-list item with its following sublist (if any)
	for i := 0; i < len(items); i++ {
		item := items[i]

		// Skip bare sublists at this level (they should only appear after text items)
		if item.IsSliceOrArray() {
			continue
		}

		escaped, _ := filterEscape(item, nil)
		sublist := ""

		// Check if the next item is a sublist for this item
		if i+1 < len(items) && items[i+1].IsSliceOrArray() {
			children := unorderedListFormatter(items[i+1], tabs+1)
			sublist = fmt.Sprintf("\n%s<ul>\n%s\n%s</ul>\n%s", indent, children, indent, indent)
			i++ // Skip the sublist item
		}

		output = append(output, fmt.Sprintf("%s<li>%s%s</li>", indent, escaped.String(), sublist))
	}

	return strings.Join(output, "\n")
}

// filterSlugify converts a string to a URL-friendly slug.
// It lowercases the string, removes non-alphanumeric characters (except hyphens and spaces),
// converts spaces to hyphens, and removes consecutive hyphens.
//
// Usage:
//
//	{{ "Hello World!"|slugify }}
//
// Output: "hello-world"
func filterSlugify(in *Value, param *Value) (*Value, error) {
	s := in.String()

	// Apply NFKD normalization to decompose accented characters into their
	// base form + combining marks (e.g., é → e + ́). This matches Django's
	// slugify behavior which uses unicodedata.normalize('NFKD') before
	// encoding to ASCII.
	s = norm.NFKD.String(s)

	s = strings.ToLower(s)

	// Replace spaces with hyphens
	s = strings.ReplaceAll(s, " ", "-")

	// Remove non-alphanumeric characters (except hyphens).
	// After NFKD normalization, combining marks (like accents) are separate
	// Unicode code points in the Mark category and will be stripped here.
	var result strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			result.WriteRune(r)
		}
	}
	s = result.String()

	// Remove consecutive hyphens
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}

	// Trim leading and trailing hyphens and underscores
	s = strings.Trim(s, "-_")

	return AsValue(s), nil
}

// filterFilesizeformat formats a file size in bytes to a human-readable string.
// Matches Django's behavior: uses non-breaking space (\u00A0) between number and unit,
// singular "byte" for ±1, and supports negative values.
//
// Usage:
//
//	{{ 123456789|filesizeformat }}
//
// Output: "117.7\u00A0MB"
func filterFilesizeformat(in *Value, param *Value) (*Value, error) {
	bytes := in.Integer()
	negative := bytes < 0
	if negative {
		bytes = -bytes
	}

	size := float64(bytes)

	units := []string{"bytes", "KB", "MB", "GB", "TB", "PB"}
	unitIdx := 0

	for size >= 1024 && unitIdx < len(units)-1 {
		size /= 1024
		unitIdx++
	}

	// Use non-breaking space (\u00A0) between number and unit (Django's avoid_wrapping)
	const nbsp = "\u00a0"
	var result string
	if unitIdx == 0 {
		// Use singular "byte" for 1 (matching Django's ngettext behavior)
		unit := "bytes"
		if bytes == 1 {
			unit = "byte"
		}
		result = fmt.Sprintf("%d%s%s", bytes, nbsp, unit)
	} else {
		result = fmt.Sprintf("%.1f%s%s", size, nbsp, units[unitIdx])
	}

	if negative {
		result = "-" + result
	}

	return AsValue(result), nil
}

// filterSafeseq applies the safe filter to each element in a sequence.
// This is useful when you have a list of strings that are known to be safe
// and want to mark each one individually.
//
// Usage:
//
//	{% for item in items|safeseq %}{{ item }}{% endfor %}
func filterSafeseq(in *Value, param *Value) (*Value, error) {
	if !in.CanSlice() {
		return in, nil
	}

	var result []*Value
	in.Iterate(func(idx, count int, key, value *Value) bool {
		var item *Value
		if value != nil {
			item = value
		} else {
			item = key
		}
		// Create a new Value marked as safe
		result = append(result, AsSafeValue(item.Interface()))
		return true
	}, func() {})

	return AsValue(result), nil
}

// filterEscapeseq applies HTML escaping to each element in a sequence.
//
// Usage:
//
//	{% for item in items|escapeseq %}{{ item }}{% endfor %}
func filterEscapeseq(in *Value, param *Value) (*Value, error) {
	if !in.CanSlice() {
		return in, nil
	}

	var result []string
	in.Iterate(func(idx, count int, key, value *Value) bool {
		var item *Value
		if value != nil {
			item = value
		} else {
			item = key
		}
		escaped, _ := filterEscape(item, nil)
		result = append(result, escaped.String())
		return true
	}, func() {})

	return AsValue(result), nil
}

// filterJSONScript safely outputs a value as JSON inside a script tag.
// The element_id argument is optional and will be used as the script tag's id.
//
// Usage:
//
//	{{ value|json_script:"my-data" }}
//	{{ value|json_script }}
//
// Output:
//
//	<script id="my-data" type="application/json">{"key":"value"}</script>
//	<script type="application/json">{"key":"value"}</script>
func filterJSONScript(in *Value, param *Value) (*Value, error) {
	var result strings.Builder

	// element_id is optional (Django 4.1+)
	if param == nil || param.IsNil() || param.String() == "" {
		result.WriteString(`<script type="application/json">`)
	} else {
		elementID := htmlEscapeReplacer.Replace(param.String())
		fmt.Fprintf(&result, `<script id="%s" type="application/json">`, elementID)
	}

	// Convert the value to JSON (json.Marshal doesn't add trailing newline)
	jsonBytes, err := json.Marshal(in.Interface())
	if err != nil {
		return nil, fmt.Errorf("json marshalling error: %w", err)
	}
	result.Write(jsonBytes)

	result.WriteString("</script>")
	return AsSafeValue(result.String()), nil
}
