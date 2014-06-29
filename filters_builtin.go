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
	RegisterFilter("safe", filterSafe)
	RegisterFilter("unsafe", filterUnsafe)
	RegisterFilter("truncatechars", filterTruncatechars)
	RegisterFilter("length", filterLength)
	RegisterFilter("upper", filterUpper)
	RegisterFilter("lower", filterLower)
	RegisterFilter("date", filterDate)
	RegisterFilter("time", filterDate) // time uses filterDate (same golang-format)
	RegisterFilter("striptags", filterStriptags)
	RegisterFilter("capfirst", filterCapfirst)

	RegisterFilter("float", filterFloat)     // pongo-specific
	RegisterFilter("integer", filterInteger) // pongo-specific

	/* Missing:
	   add
	   addslashes
	   center
	   cut
	   default
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

func filterSafe(in *Value, param *Value) (*Value, error) {
	output := strings.Replace(in.String(), "&", "&amp;", -1)
	output = strings.Replace(output, ">", "&gt;", -1)
	output = strings.Replace(output, "<", "&lt;", -1)
	return AsValue(output), nil
}

func filterUnsafe(in *Value, param *Value) (*Value, error) {
	return in, nil // nothing to do here, just to keep track of the unsafe application
}

func filterLength(in *Value, param *Value) (*Value, error) {
	return AsValue(in.Len()), nil
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
