package config

import (
	"reflect"
	"testing"
)

func TestExprParse(t *testing.T) {
	cases := []struct {
		Input  string
		Result Interpolation
		Error  bool
	}{
		{
			"foo",
			nil,
			true,
		},

		{
			`"foo"`,
			&LiteralInterpolation{Literal: "foo"},
			false,
		},

		{
			"var.foo",
			&VariableInterpolation{
				Variable: &UserVariable{
					Name: "foo",
					key:  "var.foo",
				},
			},
			false,
		},

		{
			"module.foo.bar",
			&VariableInterpolation{
				Variable: &ModuleVariable{
					Name:  "foo",
					Field: "bar",
					key:   "module.foo.bar",
				},
			},
			false,
		},

		{
			"lookup(var.foo, var.bar)",
			&FunctionInterpolation{
				Func: nil, // Funcs["lookup"]
				Args: []Interpolation{
					&VariableInterpolation{
						Variable: &UserVariable{
							Name: "foo",
							key:  "var.foo",
						},
					},
					&VariableInterpolation{
						Variable: &UserVariable{
							Name: "bar",
							key:  "var.bar",
						},
					},
				},
			},
			false,
		},

		{
			"lookup(var.foo, lookup(var.baz, var.bar))",
			&FunctionInterpolation{
				Func: nil, // Funcs["lookup"]
				Args: []Interpolation{
					&VariableInterpolation{
						Variable: &UserVariable{
							Name: "foo",
							key:  "var.foo",
						},
					},
					&FunctionInterpolation{
						Func: nil, // Funcs["lookup"]
						Args: []Interpolation{
							&VariableInterpolation{
								Variable: &UserVariable{
									Name: "baz",
									key:  "var.baz",
								},
							},
							&VariableInterpolation{
								Variable: &UserVariable{
									Name: "bar",
									key:  "var.bar",
								},
							},
						},
					},
				},
			},
			false,
		},

		{
			`concat("foo","-","0.0/16")`,
			&FunctionInterpolation{
				Func: nil, // Funcs["lookup"]
				Args: []Interpolation{
					&LiteralInterpolation{Literal: "foo"},
					&LiteralInterpolation{Literal: "-"},
					&LiteralInterpolation{Literal: "0.0/16"},
				},
			},
			false,
		},
	}

	for i, tc := range cases {
		actual, err := ExprParse(tc.Input)
		if (err != nil) != tc.Error {
			t.Fatalf("%d. Error: %s", i, err)
		}

		// This is jank, but reflect.DeepEqual never has functions
		// being the same.
		f, ok := actual.(*FunctionInterpolation)
		if ok {
			fs := make([]*FunctionInterpolation, 1)
			fs[0] = f
			for len(fs) > 0 {
				f := fs[0]
				fs = fs[1:]

				f.Func = nil
				for _, a := range f.Args {
					f, ok := a.(*FunctionInterpolation)
					if ok {
						fs = append(fs, f)
					}
				}
			}
		}

		if !reflect.DeepEqual(actual, tc.Result) {
			t.Fatalf("%d bad: %#v", i, actual)
		}
	}
}
