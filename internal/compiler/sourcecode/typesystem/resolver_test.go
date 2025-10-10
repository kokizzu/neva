package typesystem_test

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/nevalang/neva/internal/compiler/sourcecode/core"
	ts "github.com/nevalang/neva/internal/compiler/sourcecode/typesystem"
)

var errTest = errors.New("Oops!")

// TODO fix commented tests (do not remove them)
func TestExprResolver_Resolve(t *testing.T) { //nolint:maintidx
	t.Parallel()

	type testcase struct {
		expr  ts.Expr
		scope TestScope

		prepareValidator  func(v *MockexprValidatorMockRecorder)
		prepareChecker    func(c *MockcompatCheckerMockRecorder)
		prepareTerminator func(t *MockrecursionTerminatorMockRecorder)

		want    ts.Expr
		wantErr error
	}

	tests := map[string]func() testcase{
		"invalid_expr": func() testcase {
			return testcase{
				prepareValidator: func(v *MockexprValidatorMockRecorder) {
					v.Validate(ts.Expr{}).Return(errTest)
				},
				wantErr: ts.ErrInvalidExpr,
			}
		},
		"inst_expr_refers_to_undefined": func() testcase { // expr = int, scope = {}
			expr := h.Inst("int")
			return testcase{
				expr: expr,
				prepareValidator: func(v *MockexprValidatorMockRecorder) {
					v.Validate(expr).Return(nil)
				},
				scope:   TestScope{},
				wantErr: ts.ErrScope,
			}
		},
		"args_<_params": func() testcase { // expr = list<>, scope = { list<t> = list }
			expr := h.Inst("list")
			return testcase{
				expr: expr,
				scope: TestScope{
					"list": h.BaseDef(h.ParamWithNoConstr("t")),
				},
				prepareValidator: func(v *MockexprValidatorMockRecorder) {
					v.Validate(expr).Return(nil)
					v.ValidateDef(h.BaseDef(h.ParamWithNoConstr("t")))
				},
				wantErr: ts.ErrInstArgsCount,
			}
		},
		"unresolvable_argument": func() testcase { // expr = list<foo>, scope = {list<t> = list}
			expr := h.Inst("list", h.Inst("foo"))
			scope := TestScope{
				"list": h.BaseDef(ts.Param{Name: "t"}),
			}
			return testcase{
				expr:  expr,
				scope: scope,
				prepareValidator: func(v *MockexprValidatorMockRecorder) {
					v.Validate(expr).Return(nil)
					v.ValidateDef(scope["list"]).Return(nil)
					v.Validate(expr.Inst.Args[0]).Return(errTest) // in the loop
				},
				prepareTerminator: func(t *MockrecursionTerminatorMockRecorder) {
					t.ShouldTerminate(ts.NewTrace(nil, core.EntityRef{Name: "list"}), scope)
				},

				wantErr: ts.ErrUnresolvedArg,
			}
		},
		"incompat_arg": func() testcase { // expr = map<t1>, scope = { map<t t2> = map, t1 , t2 }
			expr := h.Inst("map", h.Inst("t1"))
			constr := h.Inst("t2")
			scope := TestScope{
				"map": h.BaseDef(ts.Param{"t", constr}),
				"t1":  h.BaseDef(),
				"t2":  h.BaseDef(),
			}
			return testcase{
				expr:  expr,
				scope: scope,
				prepareValidator: func(v *MockexprValidatorMockRecorder) {
					v.Validate(expr).Return(nil)
					v.ValidateDef(scope["map"]).Return(nil)
					v.Validate(h.Inst("t1")).Return(nil)
					v.ValidateDef(scope["t1"]).Return(nil)
					v.Validate(h.Inst("t2")).Return(nil)
					v.ValidateDef(scope["t2"]).Return(nil)
				},
				prepareTerminator: func(t *MockrecursionTerminatorMockRecorder) {
					t1 := ts.NewTrace(nil, core.EntityRef{Name: "map"})
					t.ShouldTerminate(t1, scope).Return(false, nil)

					t2 := ts.NewTrace(&t1, core.EntityRef{Name: "t1"})
					t.ShouldTerminate(t2, scope).Return(false, nil)

					t3 := ts.NewTrace(&t1, core.EntityRef{Name: "t2"})
					t.ShouldTerminate(t3, scope).Return(false, nil)
				},
				prepareChecker: func(c *MockcompatCheckerMockRecorder) {
					t := ts.NewTrace(nil, core.EntityRef{Name: "map"})
					tparams := ts.TerminatorParams{
						Scope:          scope,
						SubtypeTrace:   t,
						SupertypeTrace: t,
					}
					c.Check(h.Inst("t1"), h.Inst("t2"), tparams).Return(errTest)
				},
				want:    ts.Expr{},
				wantErr: ts.ErrIncompatArg,
			}
		},
		// "expr_underlaying_type_not_found": func() testcase { // expr = t1<int>, scope = { int, t1<t> = t3<t> }
		// 	expr := h.Inst("t1", h.Inst("int"))
		// 	scope := TestScope{
		// 		"int": h.BaseDef(),
		// 		"t1":  h.Def(h.Inst("t3", h.Inst("t")), h.ParamWithNoConstr("t")),
		// 	}
		// 	return testcase{
		// 		expr:  expr,
		// 		scope: scope,
		// 		prepareValidator: func(v *MockexprValidatorMockRecorder) {
		// 			v.Validate(gomock.Any()).AnyTimes().Return(nil)
		// 			v.ValidateDef(gomock.Any()).AnyTimes().Return(nil)
		// 		},
		// 		prepareTerminator: func(t *MockrecursionTerminatorMockRecorder) {
		// 			t1 := ts.NewTrace(nil, core.EntityRef{Name:"t1"})
		// 			t.ShouldTerminate(t1, scope).Return(false, nil)

		// 			t2 := ts.NewTrace(&t1, core.EntityRef{Name:"int"})
		// 			t.ShouldTerminate(t2, scope).Return(false, nil)
		// 		},
		// 		wantErr: ts.ErrScope,
		// 	}
		// },
		"constr_undefined_ref": func() testcase { // expr = t1<t2>, scope = { t2, t1<t t3> = t1 }
			expr := h.Inst("t1", h.Inst("t2"))
			constr := h.Inst("t3")
			scope := TestScope{
				"t1": h.BaseDef(ts.Param{"t", constr}),
				"t2": h.BaseDef(),
			}
			return testcase{
				expr:  expr,
				scope: scope,
				prepareValidator: func(v *MockexprValidatorMockRecorder) {
					v.Validate(gomock.Any()).AnyTimes().Return(nil)
					v.ValidateDef(gomock.Any()).AnyTimes().Return(nil)
				},
				prepareTerminator: func(t *MockrecursionTerminatorMockRecorder) {
					t1 := ts.NewTrace(nil, core.EntityRef{Name: "t1"})
					t.ShouldTerminate(t1, scope).Return(false, nil)

					t2 := ts.NewTrace(&t1, core.EntityRef{Name: "t2"})
					t.ShouldTerminate(t2, scope).Return(false, nil)
				},
				wantErr: ts.ErrConstr,
			}
		},
		"constr_ref_type_not_found": func() testcase { // expr = t1<t2>, scope = { t2, t1<t t3> }
			expr := h.Inst("t1", h.Inst("t2"))
			scope := TestScope{
				"t2": h.BaseDef(),
				"t1": h.BaseDef(h.Param("t", h.Inst("t3"))),
			}
			return testcase{
				expr:  expr,
				scope: scope,
				prepareValidator: func(v *MockexprValidatorMockRecorder) {
					v.Validate(gomock.Any()).AnyTimes().Return(nil)
					v.ValidateDef(gomock.Any()).AnyTimes().Return(nil)
				},
				prepareTerminator: func(t *MockrecursionTerminatorMockRecorder) {
					t1 := ts.NewTrace(nil, core.EntityRef{Name: "t1"})
					t.ShouldTerminate(t1, scope).Return(false, nil)
					t.ShouldTerminate(ts.NewTrace(&t1, core.EntityRef{Name: "t2"}), scope).Return(false, nil)
				},
				wantErr: ts.ErrConstr,
			}
		},
		"incompatible_arg": func() testcase { // expr = t1<t2>, scope = { t1<t t3>, t2, t3 }
			expr := h.Inst("t1", h.Inst("t2"))
			scope := TestScope{
				"t1": h.BaseDef(h.Param("t", h.Inst("t3"))),
				"t2": h.BaseDef(),
				"t3": h.BaseDef(),
			}
			return testcase{
				expr:  expr,
				scope: scope,
				prepareValidator: func(v *MockexprValidatorMockRecorder) {
					v.Validate(gomock.Any()).AnyTimes().Return(nil)
					v.ValidateDef(gomock.Any()).AnyTimes().Return(nil)
				},
				prepareChecker: func(c *MockcompatCheckerMockRecorder) {
					c.Check(h.Inst("t2"), h.Inst("t3"), gomock.Any()).Return(errTest)
				},
				prepareTerminator: func(t *MockrecursionTerminatorMockRecorder) {
					t1 := ts.NewTrace(nil, core.EntityRef{Name: "t1"})
					t.ShouldTerminate(t1, scope).Return(false, nil)

					t2 := ts.NewTrace(&t1, core.EntityRef{Name: "t2"})
					t.ShouldTerminate(t2, scope).Return(false, nil)

					t3 := ts.NewTrace(&t1, core.EntityRef{Name: "t3"})
					t.ShouldTerminate(t3, scope).Return(false, nil)
				},
				wantErr: ts.ErrIncompatArg,
			}
		},
		// Literals
		"tag-only_union": func() testcase { // expr = union{foo, bar}, scope = {}
			expr := h.Union(
				map[string]*ts.Expr{"foo": nil, "bar": nil},
			)
			return testcase{
				expr:             expr,
				prepareValidator: func(v *MockexprValidatorMockRecorder) { v.Validate(expr).Return(nil) },
				want:             expr,
				wantErr:          nil,
			}
		},
		"union_with_unresolvable_(undefined)_element": func() testcase { // t1 | t2, {t1=t1}
			t1 := h.Inst("t1")
			t2 := h.Inst("t2")
			expr := h.Union(map[string]*ts.Expr{"t1": &t1, "t2": &t2})
			scope := TestScope{"t1": h.BaseDef()} // only t1 is defined
			return testcase{
				expr:  expr,
				scope: scope,
				prepareValidator: func(v *MockexprValidatorMockRecorder) {
					v.Validate(gomock.Any()).AnyTimes().Return(nil)
					v.ValidateDef(gomock.Any()).AnyTimes().Return(nil)
				},
				prepareTerminator: func(t *MockrecursionTerminatorMockRecorder) {
					t1 := ts.NewTrace(nil, core.EntityRef{Name: "t1"})
					t.ShouldTerminate(t1, scope)

					// t2 := ts.NewTrace(nil, core.EntityRef{Name:"t2"})
					// t.ShouldTerminate(t2, scope)
				},
				wantErr: ts.ErrUnionUnresolvedEl,
			}
		},
		"union_with_unresolvable_(not_valid)_element": func() testcase { // expr = t1 | t2, scope = {t1=t1, t2=t2}
			t1 := h.Inst("t1")
			t2 := h.Inst("t2")
			expr := h.Union(map[string]*ts.Expr{"t1": &t1, "t2": &t2})
			scope := TestScope{"t1": h.BaseDef(), "t2": h.BaseDef()}
			return testcase{
				expr:  expr,
				scope: scope,
				prepareValidator: func(v *MockexprValidatorMockRecorder) {
					v.Validate(expr).Return(nil)
					v.Validate(t1).Return(nil)
					v.ValidateDef(scope["t1"]).Return(nil)
					v.Validate(t2).Return(errTest) // we make t2 invalid and thus unresolvable
				},
				prepareTerminator: func(t *MockrecursionTerminatorMockRecorder) {
					t.ShouldTerminate(
						gomock.Any(),
						gomock.Any(),
					).AnyTimes().Return(false, nil)
				},
				wantErr: ts.ErrUnionUnresolvedEl,
			}
		},
		"union_with_resolvable_elements": func() testcase { // expr = t1 | t2, scope = {t1=..., t2=...}
			t1 := h.Inst("t1")
			t2 := h.Inst("t2")
			expr := h.Union(map[string]*ts.Expr{"t1": &t1, "t2": &t2})
			scope := TestScope{"t1": h.BaseDef(), "t2": h.BaseDef()}
			return testcase{
				expr:  expr,
				scope: scope,
				prepareValidator: func(v *MockexprValidatorMockRecorder) {
					v.Validate(gomock.Any()).AnyTimes().Return(nil)
					v.ValidateDef(gomock.Any()).AnyTimes().Return(nil)
				},
				prepareTerminator: func(t *MockrecursionTerminatorMockRecorder) {
					t1 := ts.NewTrace(nil, core.EntityRef{Name: "t1"})
					t.ShouldTerminate(t1, scope)

					t2 := ts.NewTrace(nil, core.EntityRef{Name: "t2"})
					t.ShouldTerminate(t2, scope)
				},
				want: expr,
			}
		},
		"empty_record": func() testcase { // {}
			expr := h.Struct(map[string]ts.Expr{})
			scope := TestScope{}
			return testcase{
				scope: scope,
				expr:  expr,
				prepareValidator: func(v *MockexprValidatorMockRecorder) {
					v.Validate(expr).Return(nil)
				},
				want: h.Struct(map[string]ts.Expr{}),
			}
		},
		"struct_with_invalid field": func() testcase { // { name string }
			stringExpr := h.Inst("string")
			expr := h.Struct(map[string]ts.Expr{"name": stringExpr})
			scope := TestScope{}
			return testcase{
				expr:  expr,
				scope: scope,
				prepareValidator: func(v *MockexprValidatorMockRecorder) {
					v.Validate(expr).Return(nil)
					v.Validate(stringExpr).Return(errTest)
				},
				wantErr: ts.ErrRecFieldUnresolved,
			}
		},
		// "struct_with_valid_field": func() testcase { // { name string }
		// 	stringExpr := h.Inst("string")
		// 	expr := h.Struct(map[string]ts.Expr{
		// 		"name": stringExpr,
		// 	})
		// 	scope := TestScope{
		// 		"string": h.BaseDef(),
		// 	}
		// 	return testcase{
		// 		expr:  expr,
		// 		scope: scope,
		// 		prepareValidator: func(v *MockexprValidatorMockRecorder) {
		// 			v.Validate(gomock.Any()).AnyTimes().Return(nil)
		// 			v.ValidateDef(gomock.Any()).AnyTimes().Return(nil)
		// 		},
		// 		prepareTerminator: func(t *MockrecursionTerminatorMockRecorder) {
		// 			t1 := ts.NewTrace(nil, core.EntityRef{Name:"string"})
		// 			t.ShouldTerminate(t1, scope)
		// 		},
		// 		want: expr,
		// 	}
		// },
		// "param_with_same_name_as_type_in_scope_(shadowing)": func() testcase {
		// 	// t1<int>, { t1<t3>=t2<t3>, t2<t>=t3<t>, t3<t>=list<t>, list<t>, int }
		// 	scope := TestScope{
		// 		"t1": h.Def( // t1<t3> = t2<t3>
		// 			h.Inst("t2", h.Inst("t3")),
		// 			h.ParamWithNoConstr("t3"),
		// 		),
		// 		"t2": h.Def( // t2<t> = t3<t>
		// 			h.Inst("t3", h.Inst("t")),
		// 			h.ParamWithNoConstr("t"),
		// 		),
		// 		"t3": h.Def( // t3<t> = list<t>
		// 			h.Inst("list", h.Inst("t")),
		// 			h.ParamWithNoConstr("t"),
		// 		),
		// 		"list": h.BaseDef( // list<t> (base type)
		// 			h.ParamWithNoConstr("t"),
		// 		),
		// 		"int": h.BaseDef(), // int
		// 	}
		// 	return testcase{
		// 		expr:  h.Inst("t1", h.Inst("int")),
		// 		scope: scope,
		// 		prepareValidator: func(v *MockexprValidatorMockRecorder) {
		// 			v.Validate(gomock.Any()).AnyTimes().Return(nil)
		// 			v.ValidateDef(gomock.Any()).AnyTimes().Return(nil)
		// 		},
		// 		prepareTerminator: func(t *MockrecursionTerminatorMockRecorder) {
		// 			t.ShouldTerminate(gomock.Any(), gomock.Any()).AnyTimes().Return(false, nil)
		// 		},
		// 		want: h.Inst("list", h.Inst("int")),
		// 	}
		// },
		"direct_recursion_through_inst_references": func() testcase { // t, {t=t}
			scope := TestScope{
				"t": h.Def(h.Inst("t")), // direct recursion
			}
			return testcase{
				expr:  h.Inst("t"),
				scope: scope,
				prepareValidator: func(v *MockexprValidatorMockRecorder) {
					v.Validate(gomock.Any()).AnyTimes().Return(nil)
					v.ValidateDef(gomock.Any()).AnyTimes().Return(nil)
				},
				prepareTerminator: func(t *MockrecursionTerminatorMockRecorder) {
					t1 := ts.NewTrace(nil, core.EntityRef{Name: "t"})
					t.ShouldTerminate(t1, scope).Return(false, nil)

					t2 := ts.NewTrace(&t1, core.EntityRef{Name: "t"})
					t.ShouldTerminate(t2, scope).Return(false, errTest)
				},
				wantErr: ts.ErrTerminator,
			}
		},
		"indirect_(2_step)_recursion_through_inst_references": func() testcase { // t1, {t1=t2, t2=t1}
			scope := TestScope{
				"t1": h.Def(h.Inst("t2")), // indirectly
				"t2": h.Def(h.Inst("t1")), // refers to itself
			}
			return testcase{
				expr:  h.Inst("t1"),
				scope: scope,
				prepareValidator: func(v *MockexprValidatorMockRecorder) {
					v.Validate(gomock.Any()).AnyTimes().Return(nil)
					v.ValidateDef(gomock.Any()).AnyTimes().Return(nil)
				},
				prepareTerminator: func(t *MockrecursionTerminatorMockRecorder) {
					t1 := ts.NewTrace(nil, core.EntityRef{Name: "t1"})
					t.ShouldTerminate(t1, scope).Return(false, nil)

					t2 := ts.NewTrace(&t1, core.EntityRef{Name: "t2"})
					t.ShouldTerminate(t2, scope).Return(false, nil)

					t3 := ts.NewTrace(&t2, core.EntityRef{Name: "t1"})
					t.ShouldTerminate(t3, scope).Return(false, errTest)
				},
				wantErr: ts.ErrTerminator,
			}
		},
		// "substitution_of_arguments": func() testcase { // t1<int, str> { t1<p1, p2> = list<map<p1, p2>> }
		// 	scope := TestScope{
		// 		"t1": h.Def(
		// 			h.Inst(
		// 				"list",
		// 				h.Inst("map", h.Inst("p1"), h.Inst("p2")),
		// 			),
		// 			h.ParamWithNoConstr("p1"),
		// 			h.ParamWithNoConstr("p2"),
		// 		),
		// 		"int":    h.BaseDef(),
		// 		"string": h.BaseDef(),
		// 		"list":   h.BaseDef(h.ParamWithNoConstr("a")),
		// 		"map":    h.BaseDef(h.ParamWithNoConstr("a"), h.ParamWithNoConstr("b")),
		// 	}
		// 	return testcase{
		// 		expr:  h.Inst("t1", h.Inst("int"), h.Inst("string")),
		// 		scope: scope,
		// 		prepareValidator: func(v *MockexprValidatorMockRecorder) {
		// 			v.Validate(gomock.Any()).AnyTimes().Return(nil)
		// 			v.ValidateDef(gomock.Any()).AnyTimes().Return(nil)
		// 		},
		// 		prepareTerminator: func(t *MockrecursionTerminatorMockRecorder) {
		// 			t1 := ts.NewTrace(nil, core.EntityRef{Name:"t1"})
		// 			t.ShouldTerminate(t1, scope).Return(false, nil)

		// 			t2 := ts.NewTrace(&t1, core.EntityRef{Name:"int"})
		// 			t.ShouldTerminate(t2, scope).Return(false, nil)

		// 			t3 := ts.NewTrace(&t1, core.EntityRef{Name:"string"})
		// 			t.ShouldTerminate(t3, scope).Return(false, nil)

		// 			t4 := ts.NewTrace(&t1, core.EntityRef{Name:"list"})
		// 			t.ShouldTerminate(t4, scope).Return(false, nil)

		// 			t5 := ts.NewTrace(&t4, core.EntityRef{Name:"map"})
		// 			t.ShouldTerminate(t5, scope).Return(false, nil)

		// 			t6 := ts.NewTrace(&t5, core.EntityRef{Name:"p1"})
		// 			t.ShouldTerminate(t6, scope).Return(false, nil)

		// 			t7 := ts.NewTrace(&t6, core.EntityRef{Name:"int"})
		// 			t.ShouldTerminate(t7, scope).Return(false, nil)

		// 			t8 := ts.NewTrace(&t5, core.EntityRef{Name:"p2"})
		// 			t.ShouldTerminate(t8, scope).Return(false, nil)

		// 			t9 := ts.NewTrace(&t8, core.EntityRef{Name:"string"})
		// 			t.ShouldTerminate(t9, scope).Return(false, nil)
		// 		},
		// 		want: h.Inst(
		// 			"list",
		// 			h.Inst("map", h.Inst("int"), h.Inst("string")),
		// 		),
		// 	}
		// },
		// "RHS": func() testcase { // t1<int> { t1<t>=t, t=int, int }
		// 	scope := TestScope{
		// 		"t1":  h.Def(h.Inst("t"), h.ParamWithNoConstr("t")),
		// 		"t":   h.Def(h.Inst("int")),
		// 		"int": h.BaseDef(),
		// 	}
		// 	return testcase{
		// 		expr:  h.Inst("t1", h.Inst("int")),
		// 		scope: scope,
		// 		prepareValidator: func(v *MockexprValidatorMockRecorder) {
		// 			v.Validate(gomock.Any()).AnyTimes().Return(nil)
		// 			v.ValidateDef(gomock.Any()).AnyTimes().Return(nil)
		// 		},
		// 		prepareTerminator: func(t *MockrecursionTerminatorMockRecorder) {
		// 			t1 := ts.NewTrace(nil, core.EntityRef{Name:"t1"})
		// 			t.ShouldTerminate(t1, scope).Return(false, nil)

		// 			t2 := ts.NewTrace(&t1, core.EntityRef{Name:"int"})
		// 			t.ShouldTerminate(t2, scope).Return(false, nil)

		// 			t3 := ts.NewTrace(&t1, core.EntityRef{Name:"t"})
		// 			t.ShouldTerminate(t3, scope).Return(false, nil)

		// 			t4 := ts.NewTrace(&t3, core.EntityRef{Name:"int"})
		// 			t.ShouldTerminate(t4, scope).Return(false, nil)
		// 		},
		// 		want: h.Inst("int"),
		// 	}
		// },
		// "constr_refereing_type_parameter_(generics_inside_generics)": func() testcase { // t<int, list<int>> {t<a, b list<a>>, list<t>, int}
		// 	expr := h.Inst(
		// 		"t",
		// 		h.Inst("int"),
		// 		h.Inst("list", h.Inst("int")),
		// 	)
		// 	scope := TestScope{
		// 		"t": h.BaseDef(
		// 			h.ParamWithNoConstr("a"),
		// 			h.Param("b", h.Inst("list", h.Inst("a"))),
		// 		),
		// 		"list": h.BaseDef(h.ParamWithNoConstr("t")),
		// 		"int":  h.BaseDef(),
		// 	}
		// 	return testcase{
		// 		expr:  expr,
		// 		scope: scope,
		// 		prepareValidator: func(v *MockexprValidatorMockRecorder) {
		// 			v.Validate(gomock.Any()).AnyTimes()
		// 			v.ValidateDef(gomock.Any()).AnyTimes()
		// 		},
		// 		prepareTerminator: func(t *MockrecursionTerminatorMockRecorder) {
		// 			t.ShouldTerminate(gomock.Any(), gomock.Any()).
		// 				AnyTimes().
		// 				Return(false, nil)
		// 		},
		// 		prepareChecker: func(c *MockcompatCheckerMockRecorder) {
		// 			c.Check(
		// 				gomock.Any(),
		// 				gomock.Any(),
		// 				gomock.Any(),
		// 			).AnyTimes().Return(nil)
		// 		},
		// 		want: h.Inst(
		// 			"t",
		// 			h.Inst("int"),
		// 			h.Inst("list", h.Inst("int")),
		// 		),
		// 	}
		// },
		// "recursion_through_base_types_with_support_of_recursion": func() testcase { // t1 { t1 = list<t1> }
		// 	scope := TestScope{
		// 		"t1":   h.Def(h.Inst("list", h.Inst("t1"))),
		// 		"list": h.BaseDefWithRecursionAllowed(h.ParamWithNoConstr("t")),
		// 	}
		// 	return testcase{
		// 		expr:  h.Inst("t1"),
		// 		scope: scope,
		// 		prepareValidator: func(v *MockexprValidatorMockRecorder) {
		// 			v.ValidateDef(gomock.Any()).AnyTimes().Return(nil)
		// 			v.Validate(gomock.Any()).AnyTimes().Return(nil)
		// 		},
		// 		prepareTerminator: func(t *MockrecursionTerminatorMockRecorder) {
		// 			t1 := ts.NewTrace(nil, core.EntityRef{Name:"t1"})
		// 			t.ShouldTerminate(t1, scope).Return(false, nil)

		// 			t2 := ts.NewTrace(&t1, core.EntityRef{Name:"list"})
		// 			t.ShouldTerminate(t2, scope).Return(false, nil)

		// 			t3 := ts.NewTrace(&t2, core.EntityRef{Name:"t1"})
		// 			t.ShouldTerminate(t3, scope).Return(true, nil)
		// 		},
		// 		want: h.Inst("list", h.Inst("t1")),
		// 	}
		// },
		// "compatibility_check_between_two_recursive_types": func() testcase { // t3<t1> { t1 = list<t1>, t2 = list<t2>, t3<p1 t2>, list<t> }
		// 	scope := TestScope{
		// 		"t1":   h.Def(h.Inst("list", h.Inst("t1"))),
		// 		"t2":   h.Def(h.Inst("list", h.Inst("t2"))),
		// 		"t3":   h.BaseDef(h.Param("p1", h.Inst("t2"))),
		// 		"list": h.BaseDefWithRecursionAllowed(h.ParamWithNoConstr("t")),
		// 	}
		// 	return testcase{
		// 		expr:  h.Inst("t3", h.Inst("t1")),
		// 		scope: scope,
		// 		prepareValidator: func(v *MockexprValidatorMockRecorder) {
		// 			v.Validate(gomock.Any()).AnyTimes().Return(nil)
		// 			v.ValidateDef(gomock.Any()).AnyTimes().Return(nil)
		// 		},
		// 		prepareChecker: func(c *MockcompatCheckerMockRecorder) {
		// 			tparams := ts.TerminatorParams{
		// 				Scope:          scope,
		// 				SubtypeTrace:   ts.NewTrace(nil, core.EntityRef{Name:"t3"}),
		// 				SupertypeTrace: ts.NewTrace(nil, core.EntityRef{Name:"t3"}),
		// 			}
		// 			c.Check(
		// 				h.Inst("list", h.Inst("t1")),
		// 				h.Inst("list", h.Inst("t2")),
		// 				tparams,
		// 			).Return(nil)
		// 		},
		// 		prepareTerminator: func(t *MockrecursionTerminatorMockRecorder) {
		// 			t1 := ts.NewTrace(nil, core.EntityRef{Name:"t3"})
		// 			t.ShouldTerminate(t1, scope).Return(false, nil)

		// 			t2 := ts.NewTrace(&t1, core.EntityRef{Name:"t1"})
		// 			t.ShouldTerminate(t2, scope).Return(false, nil)

		// 			t3 := ts.NewTrace(&t2, core.EntityRef{Name:"list"})
		// 			t.ShouldTerminate(t3, scope).Return(false, nil)

		// 			t4 := ts.NewTrace(&t3, core.EntityRef{Name:"t1"})
		// 			t.ShouldTerminate(t4, scope).Return(true, nil)

		// 			// constr
		// 			t5 := ts.NewTrace(&t1, core.EntityRef{Name:"t2"})
		// 			t.ShouldTerminate(t5, scope).Return(false, nil)

		// 			t6 := ts.NewTrace(&t5, core.EntityRef{Name:"list"})
		// 			t.ShouldTerminate(t6, scope).Return(false, nil)

		// 			t7 := ts.NewTrace(&t6, core.EntityRef{Name:"t2"})
		// 			t.ShouldTerminate(t7, scope).Return(true, nil)
		// 		},
		// 		want: h.Inst("t3", h.Inst("list", h.Inst("t1"))),
		// 	}
		// },
	}

	for name, tt := range tests {
		name := name
		tc := tt()

		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			validator := NewMockexprValidator(ctrl)
			if tc.prepareValidator != nil {
				tc.prepareValidator(validator.EXPECT())
			}

			comparator := NewMockcompatChecker(ctrl)
			if tc.prepareChecker != nil {
				tc.prepareChecker(comparator.EXPECT())
			}

			terminator := NewMockrecursionTerminator(ctrl)
			if tc.prepareTerminator != nil {
				tc.prepareTerminator(terminator.EXPECT())
			}

			got, err := ts.MustNewResolver(validator, comparator, terminator).ResolveExpr(tc.expr, tc.scope)
			assert.Equal(t, tc.want, got)
			assert.ErrorIs(t, err, tc.wantErr)
		})
	}
}
