package irgen_test

import (
	"context"
	"testing"

	"github.com/nevalang/nevalang/internal/compiler"
	"github.com/nevalang/nevalang/internal/compiler/helper"
	"github.com/nevalang/nevalang/internal/compiler/ir"
	"github.com/nevalang/nevalang/internal/compiler/irgen"
	"github.com/stretchr/testify/assert"
)

var h helper.Helper

func TestGenerator_Generate(t *testing.T) {
	tests := []struct {
		name    string
		prog    compiler.Program
		want    ir.Program
		wantErr error
	}{
		// TODO - pkgs==nil test
		{
			name: "program that does nothing",
			prog: compiler.Program{ // start -> trigger.sigs[0]; trigger.v = 0; trigger.v -> exit
				Pkgs: map[string]compiler.Pkg{
					"io": {
						Entities: map[string]compiler.Entity{
							"Print": {
								Exported: true,
								Kind:     compiler.ComponentEntity,
								Component: compiler.Component{
									IO: compiler.IO{
										In: map[string]compiler.Port{
											"v": {},
										},
										Out: map[string]compiler.Port{
											"v": {},
										},
									},
								},
							},
						},
					},
					"flow": {
						Entities: map[string]compiler.Entity{
							"Trigger": {
								Exported: true,
								Kind:     compiler.ComponentEntity,
								Component: compiler.Component{
									IO: compiler.IO{
										In: map[string]compiler.Port{
											"sigs": {IsArr: true},
											"v":    {},
										},
										Out: map[string]compiler.Port{
											"v": {},
										},
									},
								},
							},
						},
					},
					"main": {
						Imports: h.Imports("flow"),
						Entities: map[string]compiler.Entity{
							"code": h.IntMsg(false, 0),
							"main": h.MainComponent(
								map[string]compiler.Node{
									"trigger": h.NodeWithStaticPorts(
										h.NodeInstance("flow", "Trigger", h.Rec(nil)),
										map[compiler.RelPortAddr]compiler.EntityRef{
											{Name: "v"}: {Name: "code"},
										},
									),
								},
								[]compiler.Connection{
									{
										SenderSide: compiler.SenderConnectionSide{
											PortConnectionSide: compiler.PortConnectionSide{
												PortAddr: compiler.ConnPortAddr{
													Node: "in",
													RelPortAddr: compiler.RelPortAddr{
														Name: "start",
													},
												},
											},
										},
										ReceiverSides: []compiler.PortConnectionSide{
											{
												PortAddr: compiler.ConnPortAddr{
													Node: "trigger.",
													RelPortAddr: compiler.RelPortAddr{
														Name: "sig",
													},
												},
											},
										},
									},
									{
										SenderSide: compiler.SenderConnectionSide{
											PortConnectionSide: compiler.PortConnectionSide{
												PortAddr: compiler.ConnPortAddr{
													Node: "trigger",
													RelPortAddr: compiler.RelPortAddr{
														Name: "v",
													},
												},
											},
										},
										ReceiverSides: []compiler.PortConnectionSide{
											{
												PortAddr: compiler.ConnPortAddr{
													Node: "out.exit.",
													RelPortAddr: compiler.RelPortAddr{
														Name: "sig",
													},
												},
											},
										},
									},
								},
							),
						},
					},
				},
			},
			want: ir.Program{
				Ports: map[ir.PortAddr]uint8{
					{Path: "main/in", Name: "start"}:               0,
					{Path: "main/out", Name: "exit"}:               0,
					{Path: "trigger.in", Name: "sigs", Idx: 0}:     0,
					{Path: "main/trigger/in", Name: "v", Idx: 0}:   0,
					{Path: "main/trigger/out", Name: "v", Idx: 0}:  0,
					{Path: "main/giver/out", Name: "code", Idx: 0}: 0,
				},
				Funcs: []ir.Func{
					{
						Ref: ir.FuncRef{
							Pkg:  "flow",
							Name: "Giver",
						},
						IO: ir.FuncIO{
							Out: []ir.PortAddr{
								// TODO
							},
						},
						MsgRef: "",
					},
					{
						Ref: ir.FuncRef{Pkg: "flow", Name: "Trigger"},
						IO: ir.FuncIO{
							In: []ir.PortAddr{
								{Path: "trigger.in", Name: "sigs"},
								{Path: "trigger.in", Name: "v"},
							},
							Out: []ir.PortAddr{
								{Path: "trigger.out", Name: "v"},
							},
						},
					},
				},
				Net: []ir.Connection{
					{
						SenderSide: ir.ConnectionSide{
							PortAddr: ir.PortAddr{Name: "start"},
						},
						ReceiverSides: []ir.ConnectionSide{
							{
								PortAddr: ir.PortAddr{Path: "trigger.in", Name: "sigs"},
							},
						},
					},
					{
						SenderSide: ir.ConnectionSide{
							PortAddr: ir.PortAddr{Path: "giver.out", Name: "code"},
						},
						ReceiverSides: []ir.ConnectionSide{
							{
								PortAddr: ir.PortAddr{Path: "trigger.in", Name: "v"},
							},
						},
					},
					{
						SenderSide: ir.ConnectionSide{
							PortAddr: ir.PortAddr{Path: "trigger.out", Name: "v"},
						},
						ReceiverSides: []ir.ConnectionSide{
							{
								PortAddr: ir.PortAddr{Name: "exit"},
							},
						},
					},
				},
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := irgen.New()
			got, err := g.Generate(context.Background(), tt.prog)

			assert.Equal(t, tt.want, got)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
