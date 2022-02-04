package pongo2_test

import (
	"errors"
	"github.com/flosch/pongo2/v5"
	"testing"
)

func TestVariables_Named(t *testing.T) {
	tests := map[string]struct {
		template      string
		contextObject interface{}
		want          string
		wantErr       string
	}{
		"ByReflection": {
			template:      "[{{ obj.Foo }}]",
			contextObject: testVariablesStructSimple{Foo: "someFoo"},
			want:          "[someFoo]",
		},
		"ByReflectionMethod": {
			template:      "[{{ obj.GetBar }}]",
			contextObject: testVariablesStructSimple{hiddenBar: "someBar"},
			want:          "[someBar]",
		},
		"ByReflectionFunc": {
			template:      "[{{ obj.SomeFuncVar }}]",
			contextObject: testVariablesStructSimple{SomeFuncVar: func() string { return "fromFunc" }},
			want:          "[fromFunc]",
		},
		"ByReflectionNotExported": {
			template:      "[{{ obj.hiddenBar }}]",
			contextObject: testVariablesStructSimple{hiddenBar: "someBar"},
			want:          "[]",
		},
		"ByReflectionMethodNotExported": {
			template:      "[{{ obj.hiddenGetBar }}]",
			contextObject: testVariablesStructSimple{hiddenBar: "someBar"},
			want:          "[]",
		},
		"ByNamed": {
			template:      "[{{ obj.foo }}]",
			contextObject: testVariablesStructNamed{hiddenFoo: "someFoo"},
			want:          "[someFoo]",
		},
		"ByNamedFallback": {
			template:      "[{{ obj.Bar }}]",
			contextObject: testVariablesStructNamed{Bar: "someBar"},
			want:          "[someBar]",
		},
		"ByNamedFailing": {
			template:      "[{{ obj.explode }}]",
			contextObject: testVariablesStructNamed{},
			wantErr:       "[Error (where: execution) in <string> | Line 1 Col 5 near 'obj'] can't access field explode on type struct (variable obj.explode): expected",
		},
		"ByNamedFunc": {
			template:      "[{{ obj.func }}]",
			contextObject: testVariablesStructNamed{},
			want:          "[fromFunc]",
		},
		"ByNamedAliased": {
			template:      "[{{ obj.aliased }}]",
			contextObject: testVariablesStructNamed{Aliased1: "expected"},
			want:          "[expected]",
		},
		"ByNamedAliasedConflicting": {
			template:      "[{{ obj.AliasedConflicting }}]",
			contextObject: testVariablesStructNamed{Aliased2: "expected", AliasedConflicting: "not expected, because overwritten by Aliased2"},
			want:          "[expected]",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tpl, _ := pongo2.FromString(tt.template)
			got, err := tpl.Execute(map[string]interface{}{
				"obj": tt.contextObject,
			})
			if err != nil {
				if err.Error() != tt.wantErr {
					t.Errorf("Template.Execute() error = %v, expected error: %v", err, tt.wantErr)
					return
				}
			}
			if got != tt.want {
				t.Errorf("Template.Execute() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVariables_Indexed(t *testing.T) {
	tests := map[string]struct {
		template      string
		contextObject interface{}
		want          string
		wantErr       string
	}{
		"ByReflection": {
			template:      "[{{ obj.1 }}]",
			contextObject: []string{"a", "b", "c"},
			want:          "[b]",
		},
		"ByReflectionFunc": {
			template:      "[{{ obj.0 }}]",
			contextObject: []interface{}{func() string { return "fromFunc" }},
			want:          "[fromFunc]",
		},
		"ByIndexedOnSlice": {
			template:      "[{{ obj.1 }}]",
			contextObject: testVariablesSliceIndexed{"a", "b", "c"},
			want:          "[theField]",
		},
		"ByIndexedFallbackOnSlice": {
			template:      "[{{ obj.0 }}]",
			contextObject: testVariablesSliceIndexed{"a", "b", "c"},
			want:          "[a]",
		},
		"ByIndexedFailingOnSlice": {
			template:      "[{{ obj.2 }}]",
			contextObject: testVariablesSliceIndexed{"a", "b", "c"},
			wantErr:       "[Error (where: execution) in <string> | Line 1 Col 5 near 'obj'] can't access index 2 on type slice (variable obj.2): expected",
		},
		"ByIndexedOnStruct": {
			template:      "[{{ obj.1 }}]",
			contextObject: testVariablesStructIndexed{hiddenFoo: "someFoo"},
			want:          "[someFoo]",
		},
		"ByIndexedFallbackDoesNotWorkOnStruct": {
			template:      "[{{ obj.0 }}]",
			contextObject: testVariablesStructIndexed{hiddenFoo: "someFoo"},
			wantErr:       "[Error (where: execution) in <string> | Line 1 Col 5 near 'obj'] can't access an index on type struct (variable obj.0)",
		},
		"ByIndexedFailingOnStruct": {
			template:      "[{{ obj.2 }}]",
			contextObject: testVariablesStructIndexed{hiddenFoo: "someFoo"},
			wantErr:       "[Error (where: execution) in <string> | Line 1 Col 5 near 'obj'] can't access index 2 on type struct (variable obj.2): expected",
		},
		"ByIndexedFunc": {
			template:      "[{{ obj.3 }}]",
			contextObject: testVariablesStructIndexed{},
			want:          "[fromFunc]",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tpl, _ := pongo2.FromString(tt.template)
			got, err := tpl.Execute(map[string]interface{}{
				"obj": tt.contextObject,
			})
			if err != nil {
				if err.Error() != tt.wantErr {
					t.Errorf("Template.Execute() error = %v, expected error: %v", err, tt.wantErr)
					return
				}
			}
			if got != tt.want {
				t.Errorf("Template.Execute() = %v, want %v", got, tt.want)
			}
		})
	}
}

type testVariablesStructSimple struct {
	Foo         string
	hiddenBar   string
	SomeFuncVar func() string
}

func (tv testVariablesStructSimple) GetBar() string {
	return tv.hiddenBar
}

func (tv testVariablesStructSimple) hiddenGetBar() string {
	return tv.hiddenBar
}

type testVariablesStructNamed struct {
	hiddenFoo string
	Bar       string

	Aliased1           string `pongo2:"aliased"`
	Aliased2           string `pongo2:"AliasedConflicting"`
	AliasedConflicting string
}

func (tv testVariablesStructNamed) GetNamedField(s string) (interface{}, error) {
	switch s {
	case "foo":
		return tv.hiddenFoo, nil
	case "explode":
		return nil, errTestExpected
	case "func":
		return func() string {
			return "fromFunc"
		}, nil
	default:
		return nil, pongo2.ErrNoSuchField
	}
}

type testVariablesStructIndexed struct {
	hiddenFoo string
}

func (tv testVariablesStructIndexed) GetIndexedField(s int) (interface{}, error) {
	switch s {
	case 1:
		return tv.hiddenFoo, nil
	case 2:
		return nil, errTestExpected
	case 3:
		return func() string {
			return "fromFunc"
		}, nil
	default:
		return nil, pongo2.ErrNoSuchField
	}
}

type testVariablesSliceIndexed []string

func (tv testVariablesSliceIndexed) GetIndexedField(s int) (interface{}, error) {
	switch s {
	case 1:
		return "theField", nil
	case 2:
		return nil, errTestExpected
	default:
		return nil, pongo2.ErrNoSuchField
	}
}

var (
	errTestExpected = errors.New("expected")
)
