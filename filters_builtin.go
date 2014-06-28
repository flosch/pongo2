package pongo2

import (
	"errors"
	"strconv"
	"strings"
)

func init() {
	RegisterFilter("safe", filterSafe)
	RegisterFilter("truncatechars", filterTruncatechars)
	RegisterFilter("length", filterLength)
	RegisterFilter("upper", filterUpper)
	RegisterFilter("float", filterFloat) // pongo-specific
	/* Missing:
	   add
	   addslashes
	   capfirst
	   center
	   cut
	   date
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
	   lower
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
	   striptags
	   time
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
	panic("unimplemented")
}

func filterLength(in *Value, param *Value) (*Value, error) {
	return AsValue(in.Len()), nil
}

func filterUpper(in *Value, param *Value) (*Value, error) {
	return AsValue(strings.ToUpper(in.String())), nil
}

func filterFloat(in *Value, param *Value) (*Value, error) {
	f, err := strconv.ParseFloat(in.String(), 64)
	if err != nil {
		return nil, err
	}
	return AsValue(f), nil
}
