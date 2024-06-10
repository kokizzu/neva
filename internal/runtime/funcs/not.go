package funcs

import (
	"context"

	"github.com/nevalang/neva/internal/runtime"
)

type not struct{}

func (p not) Create(io runtime.FuncIO, _ runtime.Msg) (func(ctx context.Context), error) {
	dataIn, err := io.In.Port("data")
	if err != nil {
		return nil, err
	}
	resOut, err := io.Out.Port("res")
	if err != nil {
		return nil, err
	}

	return func(ctx context.Context) {
		var val runtime.Msg

		for {

			select {
			case <-ctx.Done():
				return
			case val = <-dataIn:
			}

			select {
			case <-ctx.Done():
				return
			case resOut <- runtime.NewBoolMsg(!val.Bool()):
			}
		}
	}, nil
}
