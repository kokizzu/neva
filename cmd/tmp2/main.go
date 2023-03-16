// generator/main.go
package main

import (
	"bytes"
	"os"

	"github.com/emil14/neva/internal"
)

// pathToPkg->files
var allStdPkgsPaths = map[string]struct{}{
	"io":          {},
	"flow":        {},
	"flow/stream": {},
}
var usedStdPkgsPaths = map[string]struct{}{}

func main() {
	cleanup()

	// Tmp dir and go.mod
	if err := os.MkdirAll("tmp", os.ModePerm); err != nil {
		panic(err)
	}

	putGoMod()

	// Runtime
	if err := os.MkdirAll("tmp/internal/runtime", os.ModePerm); err != nil {
		panic(err)
	}

	runtimeBb, err := internal.RuntimeFiles.ReadFile("runtime/runtime.go")
	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer
	if _, err := buf.WriteString("// Code generated by neva. DO NOT EDIT.\n"); err != nil {
		panic(err)
	}
	if _, err := buf.Write(runtimeBb); err != nil {
		panic(err)
	}

	f, err := os.Create("tmp/internal/runtime/runtime.go")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if _, err := buf.WriteTo(f); err != nil {
		panic(err)
	}

	buf.Reset()

	// main.go
	progString := getProgString()

	f, err = os.Create("tmp/main.go")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if _, err := f.WriteString(progString); err != nil {
		panic(err)
	}

	// Std root
	if err := os.MkdirAll("tmp/internal/runtime/std/io", os.ModePerm); err != nil {
		panic(err)
	}
	bb, err := os.ReadFile("internal/runtime/std/io/io.go")
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile("tmp/internal/runtime/std/io/io.go", bb, os.ModePerm); err != nil {
		panic(err)
	}
	if err := os.MkdirAll("tmp/internal/runtime/std/flow", os.ModePerm); err != nil {
		panic(err)
	}
	bb, err = os.ReadFile("internal/runtime/std/flow/flow.go")
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile("tmp/internal/runtime/std/flow/flow.go", bb, os.ModePerm); err != nil {
		panic(err)
	}

	if err := os.Rename("tmp", "/home/evaleev/projects/tmp"); err != nil {
		panic(err)
	}

	// Move to special dir to avoid go modules problem
	// cmd := exec.Command("mv", "tmp", "/home/evaleev/projects/tmp")
	// if err := cmd.Run(); err != nil {
	// 	panic(err)
	// }

	// os.Executable()
}

func putGoMod() {
	f, err := os.Create("tmp/go.mod")
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := f.Close(); err != nil {
			panic(err)
		}
	}()

	_, err = f.WriteString("module github.com/emil14/neva")
	if err != nil {
		panic(err)
	}
}

func getProgString() string {
	return `// Code generated by neva. DO NOT EDIT.
	package main

	import (
		"context"
		"fmt"
	
		"github.com/emil14/neva/internal/runtime"
		"github.com/emil14/neva/internal/runtime/std/flow"
		"github.com/emil14/neva/internal/runtime/std/io"
	)
	
	func main() {
		// Component refs
		printerRef := runtime.ComponentRef{
			Pkg:  "io",
			Name: "printer",
		}
		triggerRef := runtime.ComponentRef{
			Pkg:  "flow",
			Name: "trigger",
		}
	
		// Routine runner
		repo := map[runtime.ComponentRef]runtime.ComponentFunc{
			printerRef: io.Print,
			triggerRef: flow.Trigger,
		}
		componentRunner := runtime.NewComponentRunner(repo)
		giverRunner := runtime.GiverRunnerImlp{}
		routineRunner := runtime.NewRoutineRunner(giverRunner, componentRunner)
	
		// Connector
		interceptor := runtime.InterceptorImlp{}
		connector := runtime.NewConnector(interceptor)
	
		// Runtime
		r := runtime.NewRuntime(connector, routineRunner)
	
		// Ports
		rootInStartPort := make(chan runtime.Msg)
		rootInStartPortAddr := runtime.PortAddr{Name: "start"}
		rootOutExitPort := make(chan runtime.Msg)
		rootOutExitPortAddr := runtime.PortAddr{Name: "exit"}
	
		printerInPort := make(chan runtime.Msg)
		printerInPortAddr := runtime.PortAddr{Path: "printer.in", Name: "v"}
	
		printerOutPort := make(chan runtime.Msg)
		printerOutPortAddr := runtime.PortAddr{Path: "printer.out", Name: "v"}
	
		triggerInSigsPort := make(chan runtime.Msg)
		triggerInSigsAddr := runtime.PortAddr{Path: "trigger.in", Name: "sigs"}
		triggerInVPort := make(chan runtime.Msg)
		triggerInVAddr := runtime.PortAddr{Path: "trigger.in", Name: "v"}
		triggerOutVPort := make(chan runtime.Msg)
		triggerOutVPortAddr := runtime.PortAddr{
			Path: "trigger.out",
			Name: "v",
		}
	
		giverOutPort := make(chan runtime.Msg)
		giverOutPortAddr := runtime.PortAddr{
			Path: "giver.out",
			Name: "code",
		}
	
		// Messages
		exitCodeOneMsg := runtime.NewIntMsg(0)
	
		prog := runtime.Program{
			Ports: map[runtime.PortAddr]chan runtime.Msg{
				// root
				rootInStartPortAddr: rootInStartPort,
				rootOutExitPortAddr: rootOutExitPort,
				// printer
				printerInPortAddr:  printerInPort,
				printerOutPortAddr: printerOutPort,
				// trigger
				triggerInSigsAddr:   triggerInSigsPort,
				triggerInVAddr:      triggerInVPort,
				triggerOutVPortAddr: triggerOutVPort,
				// giver
				giverOutPortAddr: giverOutPort,
			},
			Connections: []runtime.Connection{
				// root.start -> printer.in.v
				{
					Sender: runtime.ConnectionSide{
						Port: rootInStartPort,
						Meta: runtime.ConnectionSideMeta{
							PortAddr: rootInStartPortAddr,
						},
					},
					Receivers: []runtime.ConnectionSide{
						{
							Port: printerInPort,
							Meta: runtime.ConnectionSideMeta{
								PortAddr: printerInPortAddr,
							},
						},
					},
				},
				// printer.out.v -> trigger.in.sig
				{
					Sender: runtime.ConnectionSide{
						Port: printerOutPort,
						Meta: runtime.ConnectionSideMeta{
							PortAddr: printerOutPortAddr,
						},
					},
					Receivers: []runtime.ConnectionSide{
						{
							Port: triggerInSigsPort,
							Meta: runtime.ConnectionSideMeta{
								PortAddr: triggerInSigsAddr,
							},
						},
					},
				},
				// giver.out.code -> trigger.in.v
				{
					Sender: runtime.ConnectionSide{
						Port: giverOutPort,
						Meta: runtime.ConnectionSideMeta{
							PortAddr: giverOutPortAddr,
						},
					},
					Receivers: []runtime.ConnectionSide{
						{
							Port: triggerInVPort,
							Meta: runtime.ConnectionSideMeta{
								PortAddr: triggerInVAddr,
							},
						},
					},
				},
				// trigger.out.v -> root.out.exit
				{
					Sender: runtime.ConnectionSide{
						Port: triggerOutVPort,
						Meta: runtime.ConnectionSideMeta{
							PortAddr: triggerOutVPortAddr,
						},
					},
					Receivers: []runtime.ConnectionSide{
						{
							Port: rootOutExitPort,
							Meta: runtime.ConnectionSideMeta{
								PortAddr: rootOutExitPortAddr,
							},
						},
					},
				},
			},
			Routines: runtime.Routines{
				Giver: []runtime.GiverRoutine{
					{
						OutPort: giverOutPort,
						Msg:     exitCodeOneMsg,
					},
				},
				Component: []runtime.ComponentRoutine{
					// printer
					{
						Ref: printerRef,
						IO: runtime.IO{
							In: map[string][]chan runtime.Msg{
								"v": {printerInPort},
							},
							Out: map[string][]chan runtime.Msg{
								"v": {printerOutPort},
							},
						},
					},
					// trigger
					{
						Ref: triggerRef,
						IO: runtime.IO{
							In: map[string][]chan runtime.Msg{
								"sigs": {triggerInSigsPort},
								"v":    {triggerInVPort},
							},
							Out: map[string][]chan runtime.Msg{
								"v": {triggerOutVPort},
							},
						},
					},
				},
			},
		}
	
		exitCode, err := r.Run(context.Background(), prog)
		if err != nil {
			panic(err)
		}
	
		fmt.Println(exitCode)
	}
`
}

func cleanup() {
	if err := os.RemoveAll("tmp"); err != nil {
		panic(err)
	}
	if err := os.RemoveAll("/home/evaleev/projects/tmp"); err != nil {
		panic(err)
	}
}
