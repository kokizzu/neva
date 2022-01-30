package compiler

import (
	"errors"
	"fmt"

	"github.com/emil14/neva/internal/pkg/utils"
)

type (
	PkgManager interface {
		Pkg(string) (Pkg, error)
	}
	ModuleParser interface {
		Parse(map[string][]byte) (map[string]Module, error)
	}
	ProgramChecker interface {
		Check(Program) error
	}
	RuntimeProgramGenerator interface {
		Generate(Program) ([]byte, error)
	}
)

var (
	ErrPkgManager  = errors.New("package manager")
	ErrModParser   = errors.New("module parser")
	ErrProgChecker = errors.New("program checker")
	ErrRunProgGen  = errors.New("runtime program generator")
	ErrOpNotFound  = errors.New("operator not found")
)

type Compiler struct {
	pkg         PkgManager
	parser      ModuleParser
	checker     ProgramChecker
	generator   RuntimeProgramGenerator
	operatorsIO map[OperatorRef]IO
}

func (c Compiler) Compile(path string) ([]byte, error) {
	prog, err := c.PreCompile(path)
	if err != nil {
		return nil, fmt.Errorf("precompile: %w", err)
	}

	bb, err := c.generator.Generate(prog)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRunProgGen, err)
	}

	return bb, nil
}

func (c Compiler) PreCompile(path string) (Program, error) {
	pkg, err := c.pkg.Pkg(path)
	if err != nil {
		return Program{}, fmt.Errorf("%w: %v", ErrPkgManager, err)
	}

	mods, err := c.parser.Parse(pkg.Imports.Modules)
	if err != nil {
		return Program{}, fmt.Errorf("%w: %v", ErrModParser, err)
	}

	ops, err := c.operators(pkg.Imports.Operators)
	if err != nil {
		return Program{}, fmt.Errorf("operators: %w", err)
	}

	prog := Program{
		RootModule: pkg.RootModule,
		Scope: ProgramScope{
			Modules:   mods,
			Operators: ops,
		},
	}

	if err := c.checker.Check(prog); err != nil {
		return Program{}, fmt.Errorf("%w: %v", ErrProgChecker, err)
	}

	return prog, nil
}

func (c Compiler) operators(refs map[string]OperatorRef) (map[string]Operator, error) {
	ops := make(map[string]Operator, len(refs))

	for name, ref := range refs {
		io, ok := c.operatorsIO[ref]
		if !ok {
			return nil, fmt.Errorf("%w: %v", ErrOpNotFound, ref)
		}

		ops[name] = Operator{
			IO:  io,
			Ref: ref,
		}
	}

	return ops, nil
}

func MustNew(
	pkg PkgManager,
	parser ModuleParser,
	checker ProgramChecker,
	translator RuntimeProgramGenerator,
	operatorsIO map[OperatorRef]IO,
) Compiler {
	utils.NilArgsFatal(pkg, parser, checker, translator, operatorsIO)

	return Compiler{
		pkg:         pkg,
		parser:      parser,
		checker:     checker,
		generator:   translator,
		operatorsIO: map[OperatorRef]IO{},
	}
}
