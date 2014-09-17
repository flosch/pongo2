package pongo2

/* Filters that are provided through github.com/flosch/pongo2-addons:
   ------------------------------------------------------------------

   filesizeformat
   slugify
   timesince
   timeuntil

   Filters that won't be added:
   ----------------------------

   get_static_prefix (reason: web-framework specific)
   pprint (reason: python-specific)
   static (reason: web-framework specific)

   Reconsideration (not implemented yet):
   --------------------------------------

   force_escape (reason: not yet needed since this is the behaviour of pongo2's escape filter)
   safeseq (reason: same reason as `force_escape`)
   unordered_list (python-specific; not sure whether needed or not)
   dictsort (python-specific; maybe one could add a filter to sort a list of structs by a specific field name)
   dictsortreversed (see dictsort)
*/

import (
	"bytes"
	"errors"
	"fmt"
	"math/rand"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

func init() {
	rand.Seed(time.Now().Unix())

	RegisterFilter("escape", filterEscape)
	RegisterFilter("safe", filterSafe)
	RegisterFilter("escapejs", filterEscapejs)

	RegisterFilter("add", filterAdd)
	RegisterFilter("addslashes", filterAddslashes)
	RegisterFilter("capfirst", filterCapfirst)
	RegisterFilter("center", filterCenter)
	RegisterFilter("cut", filterCut)
	RegisterFilter("date", filterDate)
	RegisterFilter("default", filterDefault)
	RegisterFilter("default_if_none", filterDefaultIfNone)
	RegisterFilter("divisibleby", filterDivisibleby)
	RegisterFilter("first", filterFirst)
	RegisterFilter("floatformat", filterFloatformat)
	RegisterFilter("get_digit", filterGetdigit)
	RegisterFilter("iriencode", filterIriencode)
	RegisterFilter("join", filterJoin)
	RegisterFilter("last", filterLast)
	RegisterFilter("length", filterLength)
	RegisterFilter("length_is", filterLengthis)
	RegisterFilter("linebreaks", filterLinebreaks)
	RegisterFilter("linebreaksbr", filterLinebreaksbr)
	RegisterFilter("linenumbers", filterLinenumbers)
	RegisterFilter("ljust", filterLjust)
	RegisterFilter("lower", filterLower)
	RegisterFilter("make_list", filterMakelist)
	RegisterFilter("phone2numeric", filterPhone2numeric)
	RegisterFilter("pluralize", filterPluralize)
	RegisterFilter("random", filterRandom)
	RegisterFilter("removetags", filterRemovetags)
	RegisterFilter("rjust", filterRjust)
	RegisterFilter("slice", filterSlice)
	RegisterFilter("stringformat", filterStringformat)
	RegisterFilter("striptags", filterStriptags)
	RegisterFilter("time", filterDate) // time uses filterDate (same golang-format)
	RegisterFilter("title", filterTitle)
	RegisterFilter("truncatechars", filterTruncatechars)
	RegisterFilter("truncatechars_html", filterTruncatecharsHtml)
	RegisterFilter("truncatewords", filterTruncatewords)
	RegisterFilter("truncatewords_html", filterTruncatewordsHtml)
	RegisterFilter("upper", filterUpper)
	RegisterFilter("urlencode", filterUrlencode)
	RegisterFilter("urlize", filterUrlize)
	RegisterFilter("urlizetrunc", filterUrlizetrunc)
	RegisterFilter("wordcount", filterWordcount)
	RegisterFilter("wordwrap", filterWordwrap)
	RegisterFilter("yesno", filterYesno)

	RegisterFilter("float", filterFloat)     // pongo-specific
	RegisterFilter("integer", filterInteger) // pongo-specific
}

func paramsCountHelper(expected int, params []*Value) error {
	if len(params) < expected {
		return fmt.Errorf("expected at least %d parameters, but received only %d", expected, len(params))
	}
	return nil
}

func filterTruncatecharsHelper(s string, newLen int) string {
	if newLen < len(s) {
		if newLen >= 3 {
			return fmt.Sprintf("%s...", s[:newLen-3])
		}
		// Not enough space for the ellipsis
		return s[:newLen]
	}
	return s
}

func filterTruncateHtmlHelper(value string, new_output *bytes.Buffer, cond func() bool, fn func(c rune, s int, idx int) int, finalize func()) {
	vLen := len(value)
	tag_stack := make([]string, 0)
	idx := 0

	for idx < vLen && !cond() {
		c, s := utf8.DecodeRuneInString(value[idx:])
		if c == utf8.RuneError {
			idx += s
			continue
		}

		if c == '<' {
			new_output.WriteRune(c)
			idx += s // consume "<"

			if idx+1 < vLen {
				if value[idx] == '/' {
					// Close tag

					new_output.WriteString("/")

					tag := ""
					idx += 1 // consume "/"

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

					if len(tag_stack) > 0 {
						// Ideally, the close tag is TOP of tag stack
						// In malformed HTML, it must not be, so iterate through the stack and remove the tag
						for i := len(tag_stack) - 1; i >= 0; i-- {
							if tag_stack[i] == tag {
								// Found the tag
								tag_stack[i] = tag_stack[len(tag_stack)-1]
								tag_stack = tag_stack[:len(tag_stack)-1]
								break
							}
						}
					}

					new_output.WriteString(tag)
					new_output.WriteString(">")
				} else {
					// Open tag

					tag := ""

					params := false
					for idx < vLen {
						c2, size2 := utf8.DecodeRuneInString(value[idx:])
						if c2 == utf8.RuneError {
							idx += size2
							continue
						}

						new_output.WriteRune(c2)

						// End of tag found
						if c2 == '>' {
							idx++ // consume ">"
							break
						}

						if !params {
							if c2 == ' ' {
								params = true
							} else {
								tag += string(c2)
							}
						}

						idx += size2
					}

					// Add tag to stack
					tag_stack = append(tag_stack, tag)
				}
			}
		} else {
			idx = fn(c, s, idx)
		}
	}

	finalize()

	for i := len(tag_stack) - 1; i >= 0; i-- {
		tag := tag_stack[i]
		// Close everything from the regular tag stack
		new_output.WriteString(fmt.Sprintf("</%s>", tag))
	}
}

func filterTruncatechars(in *Value, params ...*Value) (*Value, error) {
	if len(params) < 1 {
		return in, nil
	}

	s := in.String()
	newLen := params[0].Integer()
	return AsValue(filterTruncatecharsHelper(s, newLen)), nil
}

func filterTruncatecharsHtml(in *Value, params ...*Value) (*Value, error) {
	if len(params) < 1 {
		return in, nil
	}

	value := in.String()
	newLen := max(params[0].Integer()-3, 0)

	new_output := bytes.NewBuffer(nil)

	textcounter := 0

	filterTruncateHtmlHelper(value, new_output, func() bool {
		return textcounter >= newLen
	}, func(c rune, s int, idx int) int {
		textcounter++
		new_output.WriteRune(c)

		return idx + s
	}, func() {
		if textcounter >= newLen && textcounter < len(value) {
			new_output.WriteString("...")
		}
	})

	return AsValue(new_output.String()), nil
}

func filterTruncatewords(in *Value, params ...*Value) (*Value, error) {
	if len(params) < 1 {
		return in, nil
	}

	words := strings.Fields(in.String())
	n := params[0].Integer()
	if n <= 0 {
		return AsValue(""), nil
	}
	nlen := min(len(words), n)
	out := make([]string, 0, nlen)
	for i := 0; i < nlen; i++ {
		out = append(out, words[i])
	}

	if n < len(words) {
		out = append(out, "...")
	}

	return AsValue(strings.Join(out, " ")), nil
}

func filterTruncatewordsHtml(in *Value, params ...*Value) (*Value, error) {
	if len(params) < 1 {
		return in, nil
	}

	value := in.String()
	newLen := max(params[0].Integer(), 0)

	new_output := bytes.NewBuffer(nil)

	wordcounter := 0

	filterTruncateHtmlHelper(value, new_output, func() bool {
		return wordcounter >= newLen
	}, func(_ rune, _ int, idx int) int {
		// Get next word
		word_found := false

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

			new_output.WriteRune(c2)
			idx += size2

			if c2 == ' ' || c2 == '.' || c2 == ',' || c2 == ';' {
				// Word ends here, stop capturing it now
				break
			} else {
				word_found = true
			}
		}

		if word_found {
			wordcounter++
		}

		return idx
	}, func() {
		if wordcounter >= newLen {
			new_output.WriteString("...")
		}
	})

	return AsValue(new_output.String()), nil
}

func filterEscape(in *Value, params ...*Value) (*Value, error) {
	output := strings.Replace(in.String(), "&", "&amp;", -1)
	output = strings.Replace(output, ">", "&gt;", -1)
	output = strings.Replace(output, "<", "&lt;", -1)
	output = strings.Replace(output, "\"", "&quot;", -1)
	output = strings.Replace(output, "'", "&#39;", -1)
	return AsValue(output), nil
}

func filterSafe(in *Value, params ...*Value) (*Value, error) {
	return in, nil // nothing to do here, just to keep track of the safe application
}

func filterEscapejs(in *Value, params ...*Value) (*Value, error) {
	sin := in.String()

	var b bytes.Buffer

	idx := 0
	for idx < len(sin) {
		c, size := utf8.DecodeRuneInString(sin[idx:])
		if c == utf8.RuneError {
			idx += size
			continue
		}

		if c == '\\' {
			// Escape seq?
			if idx+1 < len(sin) {
				switch sin[idx+1] {
				case 'r':
					b.WriteString(fmt.Sprintf(`\u%04X`, '\r'))
					idx += 2
					continue
				case 'n':
					b.WriteString(fmt.Sprintf(`\u%04X`, '\n'))
					idx += 2
					continue
					/*case '\'':
						b.WriteString(fmt.Sprintf(`\u%04X`, '\''))
						idx += 2
						continue
					case '"':
						b.WriteString(fmt.Sprintf(`\u%04X`, '"'))
						idx += 2
						continue*/
				}
			}
		}

		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == ' ' || c == '/' {
			b.WriteRune(c)
		} else {
			b.WriteString(fmt.Sprintf(`\u%04X`, c))
		}

		idx += size
	}

	return AsValue(b.String()), nil
}

func filterAdd(in *Value, params ...*Value) (*Value, error) {
	if len(params) < 1 {
		return in, nil
	}

	param := params[0]
	if in.IsNumber() && param.IsNumber() {
		if in.IsFloat() || param.IsFloat() {
			return AsValue(in.Float() + param.Float()), nil
		} else {
			return AsValue(in.Integer() + param.Integer()), nil
		}
	}
	// If in/param is not a number, we're relying on the
	// Value's String() convertion and just add them both together
	return AsValue(in.String() + param.String()), nil
}

func filterAddslashes(in *Value, params ...*Value) (*Value, error) {
	output := strings.Replace(in.String(), "\\", "\\\\", -1)
	output = strings.Replace(output, "\"", "\\\"", -1)
	output = strings.Replace(output, "'", "\\'", -1)
	return AsValue(output), nil
}

func filterCut(in *Value, params ...*Value) (*Value, error) {
	if len(params) < 1 {
		return in, nil
	}
	return AsValue(strings.Replace(in.String(), params[0].String(), "", -1)), nil
}

func filterLength(in *Value, params ...*Value) (*Value, error) {
	return AsValue(in.Len()), nil
}

func filterLengthis(in *Value, params ...*Value) (*Value, error) {
	if len(params) < 1 {
		return AsValue(false), nil
	}
	return AsValue(in.Len() == params[0].Integer()), nil
}

func filterDefault(in *Value, params ...*Value) (*Value, error) {
	if !in.IsTrue() {
		if len(params) < 1 {
			return AsValue(""), nil
		}
		return params[0], nil
	}
	return in, nil
}

func filterDefaultIfNone(in *Value, params ...*Value) (*Value, error) {
	if in.IsNil() {
		if len(params) < 1 {
			return AsValue(""), nil
		}
		return params[0], nil
	}
	return in, nil
}

func filterDivisibleby(in *Value, params ...*Value) (*Value, error) {
	if len(params) < 1 || params[0].Integer() == 0 {
		return AsValue(false), nil
	}
	return AsValue(in.Integer()%params[0].Integer() == 0), nil
}

func filterFirst(in *Value, params ...*Value) (*Value, error) {
	if in.CanSlice() && in.Len() > 0 {
		return in.Index(0), nil
	}
	return AsValue(""), nil
}

func filterFloatformat(in *Value, params ...*Value) (*Value, error) {
	val := in.Float()

	decimals := -1
	if len(params) > 0 && !params[0].IsNil() {
		// Any argument provided?
		decimals = params[0].Integer()
	}

	// if the argument is not a number (e. g. empty), the default
	// behaviour is trim the result
	trim := len(params) > 0 && !params[0].IsNumber()

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

	return AsValue(strconv.FormatFloat(val, 'f', decimals, 64)), nil
}

func filterGetdigit(in *Value, params ...*Value) (*Value, error) {
	if len(params) < 1 {
		return in, nil
	}
	i := params[0].Integer()
	l := len(in.String()) // do NOT use in.Len() here!
	if i <= 0 || i > l {
		return in, nil
	}
	return AsValue(in.String()[l-i] - 48), nil
}

const filterIRIChars = "/#%[]=:;$&()+,!?*@'~"

func filterIriencode(in *Value, params ...*Value) (*Value, error) {
	var b bytes.Buffer

	sin := in.String()
	for _, r := range sin {
		if strings.IndexRune(filterIRIChars, r) >= 0 {
			b.WriteRune(r)
		} else {
			b.WriteString(url.QueryEscape(string(r)))
		}
	}

	return AsValue(b.String()), nil
}

func filterJoin(in *Value, params ...*Value) (*Value, error) {
	if !in.CanSlice() {
		return in, nil
	}

	sl := make([]string, 0, in.Len())
	for i := 0; i < in.Len(); i++ {
		sl = append(sl, in.Index(i).String())
	}

	if len(params) < 1 {
		return AsValue(strings.Join(sl, "")), nil
	}

	return AsValue(strings.Join(sl, params[0].String())), nil
}

func filterLast(in *Value, params ...*Value) (*Value, error) {
	if in.CanSlice() && in.Len() > 0 {
		return in.Index(in.Len() - 1), nil
	}
	return AsValue(""), nil
}

func filterUpper(in *Value, params ...*Value) (*Value, error) {
	return AsValue(strings.ToUpper(in.String())), nil
}

func filterLower(in *Value, params ...*Value) (*Value, error) {
	return AsValue(strings.ToLower(in.String())), nil
}

func filterMakelist(in *Value, params ...*Value) (*Value, error) {
	s := in.String()
	result := make([]string, 0, len(s))
	for _, c := range s {
		result = append(result, string(c))
	}
	return AsValue(result), nil
}

func filterCapfirst(in *Value, params ...*Value) (*Value, error) {
	if in.Len() <= 0 {
		return AsValue(""), nil
	}
	t := in.String()
	return AsValue(strings.ToUpper(string(t[0])) + t[1:]), nil
}

func filterCenter(in *Value, params ...*Value) (*Value, error) {
	if len(params) < 1 {
		return in, nil
	}

	width := params[0].Integer()
	slen := in.Len()
	if width <= slen {
		return in, nil
	}

	spaces := width - slen
	left := spaces/2 + spaces%2
	right := spaces / 2

	return AsValue(fmt.Sprintf("%s%s%s", strings.Repeat(" ", left),
		in.String(), strings.Repeat(" ", right))), nil
}

func filterDate(in *Value, params ...*Value) (*Value, error) {
	t, is_time := in.Interface().(time.Time)
	if !is_time {
		return nil, errors.New("Filter input argument must be of type 'time.Time'.")
	}

	if len(params) < 1 {
		return AsValue(t.Format(time.RFC822)), nil
	}

	return AsValue(t.Format(params[0].String())), nil
}

func filterFloat(in *Value, params ...*Value) (*Value, error) {
	return AsValue(in.Float()), nil
}

func filterInteger(in *Value, params ...*Value) (*Value, error) {
	return AsValue(in.Integer()), nil
}

func filterLinebreaks(in *Value, params ...*Value) (*Value, error) {
	if in.Len() == 0 {
		return in, nil
	}

	var b bytes.Buffer

	// Newline = <br />
	// Double newline = <p>...</p>
	lines := strings.Split(in.String(), "\n")
	lenlines := len(lines)

	opened := false

	for idx, line := range lines {

		if !opened {
			b.WriteString("<p>")
			opened = true
		}

		b.WriteString(line)

		if idx < lenlines-1 && strings.TrimSpace(lines[idx]) != "" {
			// We've not reached the end
			if strings.TrimSpace(lines[idx+1]) == "" {
				// Next line is empty
				if opened {
					b.WriteString("</p>")
					opened = false
				}
			} else {
				b.WriteString("<br />")
			}
		}
	}

	if opened {
		b.WriteString("</p>")
	}

	return AsValue(b.String()), nil
}

func filterLinebreaksbr(in *Value, params ...*Value) (*Value, error) {
	return AsValue(strings.Replace(in.String(), "\n", "<br />", -1)), nil
}

func filterLinenumbers(in *Value, params ...*Value) (*Value, error) {
	lines := strings.Split(in.String(), "\n")
	output := make([]string, 0, len(lines))
	for idx, line := range lines {
		output = append(output, fmt.Sprintf("%d. %s", idx+1, line))
	}
	return AsValue(strings.Join(output, "\n")), nil
}

func filterLjust(in *Value, params ...*Value) (*Value, error) {
	times := 0
	if len(params) > 0 {
		times = params[0].Integer() - in.Len()
		if times < 0 {
			times = 0
		}
	}

	return AsValue(fmt.Sprintf("%s%s", in.String(), strings.Repeat(" ", times))), nil
}

func filterUrlencode(in *Value, params ...*Value) (*Value, error) {
	return AsValue(url.QueryEscape(in.String())), nil
}

// TODO: This regexp could do some work
var filterUrlizeURLRegexp = regexp.MustCompile(`((((http|https)://)|www\.|((^|[ ])[0-9A-Za-z_\-]+(\.com|\.net|\.org|\.info|\.biz|\.de))))(?U:.*)([ ]+|$)`)
var filterUrlizeEmailRegexp = regexp.MustCompile(`(\w+@\w+\.\w{2,4})`)

func filterUrlizeHelper(input string, autoescape bool, trunc int) string {
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

		t, err := ApplyFilter("iriencode", AsValue(raw_url))
		if err != nil {
			panic(err)
		}
		url := t.String()

		if !strings.HasPrefix(url, "http") {
			url = fmt.Sprintf("http://%s", url)
		}

		title := raw_url

		if trunc > 3 && len(title) > trunc {
			title = fmt.Sprintf("%s...", title[:trunc-3])
		}

		if autoescape {
			t, err := ApplyFilter("escape", AsValue(title))
			if err != nil {
				panic(err)
			}
			title = t.String()
		}

		return fmt.Sprintf(`%s<a href="%s" rel="nofollow">%s</a>%s`, prefix, url, title, suffix)
	})

	sout = filterUrlizeEmailRegexp.ReplaceAllStringFunc(sout, func(mail string) string {

		title := mail

		if trunc > 3 && len(title) > trunc {
			title = fmt.Sprintf("%s...", title[:trunc-3])
		}

		return fmt.Sprintf(`<a href="mailto:%s">%s</a>`, mail, title)
	})

	return sout
}

func filterUrlize(in *Value, params ...*Value) (*Value, error) {
	autoescape := true
	if len(params) > 0 && params[0].IsBool() {
		autoescape = params[0].Bool()
	}

	return AsValue(filterUrlizeHelper(in.String(), autoescape, -1)), nil
}

func filterUrlizetrunc(in *Value, params ...*Value) (*Value, error) {
	if err := paramsCountHelper(1, params); err != nil {
		return AsValue(""), err
	}
	return AsValue(filterUrlizeHelper(in.String(), true, params[0].Integer())), nil
}

func filterStringformat(in *Value, params ...*Value) (*Value, error) {
	if len(params) < 1 {
		return AsValue(fmt.Sprintf("%v", in.Interface())), nil
	}
	return AsValue(fmt.Sprintf(params[0].String(), in.Interface())), nil
}

var re_striptags = regexp.MustCompile("<[^>]*?>")

func filterStriptags(in *Value, params ...*Value) (*Value, error) {
	s := in.String()

	// Strip all tags
	s = re_striptags.ReplaceAllString(s, "")

	return AsValue(strings.TrimSpace(s)), nil
}

// https://en.wikipedia.org/wiki/Phoneword
var filterPhone2numericMap = map[string]string{
	"a": "2", "b": "2", "c": "2", "d": "3", "e": "3", "f": "3", "g": "4", "h": "4", "i": "4", "j": "5", "k": "5",
	"l": "5", "m": "6", "n": "6", "o": "6", "p": "7", "q": "7", "r": "7", "s": "7", "t": "8", "u": "8", "v": "8",
	"w": "9", "x": "9", "y": "9", "z": "9",
}

func filterPhone2numeric(in *Value, params ...*Value) (*Value, error) {
	sin := in.String()
	for k, v := range filterPhone2numericMap {
		sin = strings.Replace(sin, k, v, -1)
		sin = strings.Replace(sin, strings.ToUpper(k), v, -1)
	}
	return AsValue(sin), nil
}

func filterPluralize(in *Value, params ...*Value) (*Value, error) {
	if in.IsNumber() {
		if len(params) > 0 {
			// Old style (extract data)
			if len(params) == 1 && strings.Contains(params[0].String(), ","){
				endings := strings.Split(params[0].String(), ",")
				params = []*Value{}
				for _, endV := range endings {
					params = append(params, AsValue(endV))
				}
			}

			// New Style
			if len(params) > 2 {
				return nil, errors.New("You cannot pass more than 2 arguments to filter 'pluralize'.")
			}

			if len(params) == 1 {
				// 1 argument
				if in.Integer() != 1 {
					return params[0], nil
				}
			} else {
				if in.Integer() != 1 {
					// 2 arguments
					return params[1], nil
				}
				return params[0], nil
			}
		} else {
			if in.Integer() != 1 {
				// return default 's'
				return AsValue("s"), nil
			}
		}
		return AsValue(""), nil
	} else {
		return nil, errors.New("Filter 'pluralize' does only work on numbers.")
	}
}

func filterRandom(in *Value, params ...*Value) (*Value, error) {
	if !in.CanSlice() || in.Len() <= 0 {
		return in, nil
	}
	i := rand.Intn(in.Len())
	return in.Index(i), nil
}

func filterRemovetags(in *Value, params ...*Value) (*Value, error) {
	if len(params) < 1 {
		return in, nil
	}

	s := in.String()
	tags := strings.Split(params[0].String(), ",")

	// Strip only specific tags
	for _, tag := range tags {
		re := regexp.MustCompile(fmt.Sprintf("</?%s/?>", tag))
		s = re.ReplaceAllString(s, "")
	}

	return AsValue(strings.TrimSpace(s)), nil
}

func filterRjust(in *Value, params ...*Value) (*Value, error) {
	times := 0
	if len(params) > 0 {
		times = params[0].Integer()
	}
	return AsValue(fmt.Sprintf(fmt.Sprintf("%%%ds", times), in.String())), nil
}

func filterSlice(in *Value, params ...*Value) (*Value, error) {
	if !in.CanSlice() {
		return in, nil
	}
	if err := paramsCountHelper(1, params); err != nil {
		return nil, err
	}

	// Old style (extract data)
	if len(params) == 1 && strings.Contains(params[0].String(), ":"){
		values := strings.Split(params[0].String(), ":")
		params = []*Value{}
		for _, v := range values {
			params = append(params, AsValue(v))
		}
	}

	if len(params) != 2 {
		return nil, errors.New("Slice string must have the format 'from:to' [from/to can be omitted, but the ':' is required]")
	}

	from := params[0].Integer()
	to := in.Len()

	if from > to {
		from = to
	}

	vto := params[1].Integer()
	if vto >= from && vto <= in.Len() {
		to = vto
	}

	return in.Slice(from, to), nil
}

func filterTitle(in *Value, params ...*Value) (*Value, error) {
	if !in.IsString() {
		return AsValue(""), nil
	}
	return AsValue(strings.Title(strings.ToLower(in.String()))), nil
}

func filterWordcount(in *Value, params ...*Value) (*Value, error) {
	return AsValue(len(strings.Fields(in.String()))), nil
}

func filterWordwrap(in *Value, params ...*Value) (*Value, error) {
	if len(params) < 1 {
		return in, nil
	}

	words := strings.Fields(in.String())
	words_len := len(words)
	wrap_at := params[0].Integer()
	if wrap_at <= 0 {
		return in, nil
	}

	linecount := words_len/wrap_at + words_len%wrap_at
	lines := make([]string, 0, linecount)
	for i := 0; i < linecount; i++ {
		lines = append(lines, strings.Join(words[wrap_at*i:min(wrap_at*(i+1), words_len)], " "))
	}
	return AsValue(strings.Join(lines, "\n")), nil
}

func filterYesno(in *Value, params ...*Value) (*Value, error) {
	choices := map[int]string{
		0: "yes",
		1: "no",
		2: "maybe",
	}

	// Old style (extract data)
	if len(params) == 1 && strings.Contains(params[0].String(), ","){
		customChoices := strings.Split(params[0].String(), ",")
		params = []*Value{}
		for _, chV := range customChoices {
			params = append(params, AsValue(chV))
		}
	}

	if len(params) > 0 {
		if len(params) > 3 {
			return nil, fmt.Errorf("You cannot pass more than 3 options to the 'yesno'-filter (got %d: '%v').", len(params), params)
		}
		if len(params) < 2 {
			return nil, fmt.Errorf("You must pass either no or at least 2 arguments to the 'yesno'-filter (got: '%v').", params)
		}

		// Map to the options now
		for i := 0; i < len(params); i++ {
			if !params[i].IsNil() {
				choices[i] = params[i].String()
			}
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
