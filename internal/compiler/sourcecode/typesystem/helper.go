package typesystem

import "github.com/nevalang/neva/internal/compiler/sourcecode/core"

// Helper is just a namespace for helper functions to avoid conflicts with entity types.
// It's a stateless type and it's safe to share it between goroutines.
type Helper struct{}

// any base type def (without body) that has type parameters allows recursion
func (h Helper) BaseDefWithRecursionAllowed(params ...Param) Def {
	return Def{
		Params:   params,
		BodyExpr: nil,
	}
}

// Do not pass empty string as a name to avoid Body.Empty() == true
func (h Helper) BaseDef(params ...Param) Def {
	return Def{Params: params}
}

func (h Helper) Def(body Expr, params ...Param) Def {
	return Def{
		Params:   params,
		BodyExpr: &body,
	}
}

// Do not pass empty string as a name to avoid inst.Empty() == true
func (h Helper) Inst(ref string, args ...Expr) Expr {
	if args == nil {
		args = []Expr{} // makes easier testing because resolver returns non-nil args for native types
	}
	return Expr{
		Inst: &InstExpr{
			Ref:  core.EntityRef{Name: ref},
			Args: args,
		},
	}
}

func (h Helper) Union(els map[string]*Expr) Expr {
	if els == nil { // for !lit.Empty()
		els = map[string]*Expr{}
	}
	return Expr{
		Lit: &LitExpr{Union: els},
	}
}

func (h Helper) Struct(structure map[string]Expr) Expr {
	if structure == nil { // for !lit.Empty()
		structure = map[string]Expr{}
	}
	return Expr{
		Lit: &LitExpr{
			Struct: structure,
		},
	}
}

func (h Helper) ParamWithNoConstr(name string) Param {
	return Param{
		Name: name,
		Constr: Expr{
			Inst: &InstExpr{
				Ref: core.EntityRef{Name: "any"},
			},
		},
	}
}

func (h Helper) Param(name string, constr Expr) Param {
	return Param{
		Name:   name,
		Constr: constr,
	}
}

type DefaultStringer string

func (ds DefaultStringer) String() string { return string(ds) }

func (h Helper) Trace(ss ...string) Trace {
	var t *Trace
	for _, s := range ss {
		tmp := NewTrace(t, core.EntityRef{Name: s})
		t = &tmp
	}
	return *t
}
