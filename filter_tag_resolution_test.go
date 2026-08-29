package pongo2_test

import (
	"strings"
	"testing"

	"github.com/flosch/pongo2/v7"
)

func TestFilterTagResolvesFiltersAtParseTime(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		filter string
		setup  func(*pongo2.TemplateSet) error
		want   string
	}{
		{
			name:   "banned",
			filter: "upper",
			setup:  func(set *pongo2.TemplateSet) error { return set.BanFilter("upper") },
			want:   "Usage of filter 'upper' is not allowed",
		},
		{
			name:   "unknown",
			filter: "does_not_exist",
			setup:  func(*pongo2.TemplateSet) error { return nil },
			want:   "Filter 'does_not_exist' does not exist",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			set := pongo2.NewSet(test.name, &pongo2.DummyLoader{})
			if err := test.setup(set); err != nil {
				t.Fatalf("setup: %v", err)
			}
			_, err := set.FromString("{% filter " + test.filter + " %}body{% endfilter %}")
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("FromString error = %v, want %q", err, test.want)
			}
		})
	}
}
