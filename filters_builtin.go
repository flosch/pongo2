package pongo2

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func init() {
	RegisterFilter("escape", filterEscape)
	RegisterFilter("safe", filterSafe)

	RegisterFilter("add", filterAdd)
	RegisterFilter("capfirst", filterCapfirst)
	RegisterFilter("cut", filterCut)
	RegisterFilter("date", filterDate)
	RegisterFilter("default", filterDefault)
	RegisterFilter("default_if_none", filterDefaultIfNone)
	RegisterFilter("length", filterLength)
	RegisterFilter("lower", filterLower)
	RegisterFilter("upper", filterUpper)
	RegisterFilter("striptags", filterStriptags)
	RegisterFilter("time", filterDate) // time uses filterDate (same golang-format)
	RegisterFilter("truncatechars", filterTruncatechars)

	RegisterFilter("float", filterFloat)     // pongo-specific
	RegisterFilter("integer", filterInteger) // pongo-specific

	/* Missing:
	   addslashes
	   center
	   default_if_none
	   dictsort
	   dictsortreversed
	   divisibleby
	   escape
	   escapejs
	   filesizeformat
	   first
	   floatformat
	   force_escape
	   get_digit
	   iriencode
	   join
	   last
	   length_is
	   linebreaks
	   linebreaksbr
	   linenumbers
	   ljust
	   make_list
	   phone2numeric
	   pluralize
	   pprint
	   random
	   removetags
	   rjust
	   safeseq
	   slice
	   slugify
	   stringformat
	   timesince
	   timeuntil
	   title
	   truncatechars_html
	   truncatewords
	   truncatewords_html
	   unordered_list
	   urlencode
	   urlize
	   urlizetrunc
	   wordcount
	   wordwrap
	   yesno
	*/
}

func filterTruncatechars(in *Value, param *Value) (*Value, error) {
	if !in.CanSlice() {
		return nil, errors.New("")
	}
	return in.Slice(0, param.Integer()), nil
}

func filterEscape(in *Value, param *Value) (*Value, error) {
	output := strings.Replace(in.String(), "&", "&amp;", -1)
	output = strings.Replace(output, ">", "&gt;", -1)
	output = strings.Replace(output, "<", "&lt;", -1)
	output = strings.Replace(output, "\"", "&quot;", -1)
	output = strings.Replace(output, "'", "&#39;", -1)
	return AsValue(output), nil
}

func filterSafe(in *Value, param *Value) (*Value, error) {
	return in, nil // nothing to do here, just to keep track of the safe application
}

func filterAdd(in *Value, param *Value) (*Value, error) {
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

func filterCut(in *Value, param *Value) (*Value, error) {
	return AsValue(strings.Replace(in.String(), param.String(), "", -1)), nil
}

func filterLength(in *Value, param *Value) (*Value, error) {
	return AsValue(in.Len()), nil
}

func filterDefault(in *Value, param *Value) (*Value, error) {
	if !in.IsTrue() {
		return param, nil
	}
	return in, nil
}

func filterDefaultIfNone(in *Value, param *Value) (*Value, error) {
	if in.IsNil() {
		return param, nil
	}
	return in, nil
}

func filterUpper(in *Value, param *Value) (*Value, error) {
	return AsValue(strings.ToUpper(in.String())), nil
}

func filterLower(in *Value, param *Value) (*Value, error) {
	return AsValue(strings.ToLower(in.String())), nil
}

func filterCapfirst(in *Value, param *Value) (*Value, error) {
	return AsValue(strings.Title(in.String())), nil
}

func filterDate(in *Value, param *Value) (*Value, error) {
	t, is_time := in.Interface().(time.Time)
	if !is_time {
		return nil, errors.New("Filter input argument must be of type 'time.Time'.")
	}
	return AsValue(t.Format(param.String())), nil
}

func filterFloat(in *Value, param *Value) (*Value, error) {
	f, err := strconv.ParseFloat(in.String(), 64)
	if err != nil {
		return nil, err
	}
	return AsValue(f), nil
}

func filterInteger(in *Value, param *Value) (*Value, error) {
	i, err := strconv.Atoi(in.String())
	if err != nil {
		return nil, err
	}
	return AsValue(i), nil
}

var re_striptags = regexp.MustCompile("<[^>]*?>")

func filterStriptags(in *Value, param *Value) (*Value, error) {
	s := in.String()
	tags := strings.Split(param.String(), ",")
	if param.Len() > 0 && len(tags) > 0 {
		// Strip only specific tags
		for _, tag := range tags {
			re := regexp.MustCompile(fmt.Sprintf("</?%s/?>", tag))
			s = re.ReplaceAllString(s, "")
		}
	} else {
		// Strip all tags
		s = re_striptags.ReplaceAllString(s, "")
	}
	return AsValue(strings.TrimSpace(s)), nil
}
