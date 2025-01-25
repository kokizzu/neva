package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/nevalang/neva/internal/compiler"

	cli "github.com/urfave/cli/v2"
)

func newRunCmd(workdir string, nativec compiler.Compiler) *cli.Command {
	return &cli.Command{
		Name:  "run",
		Usage: "Build and run neva program from source code",
		Args:  true,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "trace",
				Usage: "Write trace information to file",
			},
		},
		ArgsUsage: "Provide path to main package",
		Action: func(cliCtx *cli.Context) error {
			mainPkg, err := mainPkgPathFromArgs(cliCtx)
			if err != nil {
				return err
			}

			var trace bool
			if cliCtx.IsSet("trace") {
				trace = true
			}

			input := compiler.CompilerInput{
				Main:   mainPkg,
				Output: workdir,
				Trace:  trace,
			}

			if err := nativec.Compile(cliCtx.Context, input); err != nil {
				return err
			}

			// here we're making assumptions about compiler internals
			expectedOutputFileName := "output"
			if runtime.GOOS == "windows" {
				expectedOutputFileName += ".exe"
			}

			execPath := filepath.Join(workdir, expectedOutputFileName)

			defer func() {
				if err := os.Remove(execPath); err != nil {
					fmt.Println("failed to remove output file:", err)
				}
			}()

			cmd := exec.CommandContext(cliCtx.Context, execPath)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to run generated executable: %w", err)
			}

			return nil
		},
	}
}
