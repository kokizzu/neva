package compiler

import (
	"context"
	"strings"

	"github.com/nevalang/neva/internal/compiler/ir"
	"github.com/nevalang/neva/internal/compiler/sourcecode"
)

type Compiler struct {
	fe Frontend
	me Middleend
	be Backend
}

type CompilerInput struct {
	Main   string
	Output string
	Trace  bool
}

func (c Compiler) Compile(ctx context.Context, input CompilerInput) error {
	feResult, err := c.fe.Process(ctx, input.Main)
	if err != nil {
		return err
	}

	meResult, err := c.me.Process(feResult)
	if err != nil {
		return err
	}

	return c.be.Emit(input.Output, meResult.IR, input.Trace)
}

type Frontend struct {
	builder Builder
	parser  Parser
}

type FrontendResult struct {
	MainPkg     string
	RawBuild    RawBuild
	ParsedBuild sourcecode.Build
}

func (f Frontend) Process(ctx context.Context, main string) (FrontendResult, *Error) {
	raw, root, err := f.builder.Build(ctx, main)
	if err != nil {
		return FrontendResult{}, Error{Location: &sourcecode.Location{Package: main}}.Wrap(err)
	}

	parsedMods, err := f.parser.ParseModules(raw.Modules)
	if err != nil {
		return FrontendResult{}, err
	}

	parsedBuild := sourcecode.Build{
		EntryModRef: raw.EntryModRef,
		Modules:     parsedMods,
	}

	mainPkg := strings.TrimPrefix(main, "./")
	mainPkg = strings.TrimPrefix(mainPkg, root+"/")

	return FrontendResult{
		ParsedBuild: parsedBuild,
		RawBuild:    raw,
		MainPkg:     mainPkg,
	}, nil
}

func NewFrontend(builder Builder, parser Parser) Frontend {
	return Frontend{
		builder: builder,
		parser:  parser,
	}
}

type Middleend struct {
	desugarer Desugarer
	analyzer  Analyzer
	irgen     IRGenerator
}

type MiddleendResult struct {
	AnalyzedBuild  sourcecode.Build
	DesugaredBuild sourcecode.Build
	IR             *ir.Program
}

func (m Middleend) Process(feResult FrontendResult) (MiddleendResult, *Error) {
	analyzedBuild, err := m.analyzer.AnalyzeExecutableBuild(
		feResult.ParsedBuild,
		feResult.MainPkg,
	)
	if err != nil {
		return MiddleendResult{}, err
	}

	desugaredBuild, err := m.desugarer.Desugar(analyzedBuild)
	if err != nil {
		return MiddleendResult{}, err
	}

	irProg, err := m.irgen.Generate(desugaredBuild, feResult.MainPkg)
	if err != nil {
		return MiddleendResult{}, err
	}

	return MiddleendResult{
		AnalyzedBuild:  analyzedBuild,
		DesugaredBuild: desugaredBuild,
		IR:             irProg,
	}, nil
}

func New(
	builder Builder,
	parser Parser,
	desugarer Desugarer,
	analyzer Analyzer,
	irgen IRGenerator,
	backend Backend,
) Compiler {
	return Compiler{
		fe: Frontend{
			builder: builder,
			parser:  parser,
		},
		me: Middleend{
			desugarer: desugarer,
			analyzer:  analyzer,
			irgen:     irgen,
		},
		be: backend,
	}
}
