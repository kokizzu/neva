package funcs

import (
	"context"
	"sync"

	"github.com/nevalang/neva/internal/runtime"
)

type stringAdd struct{}

func (stringAdd) Create(
	io runtime.IO,
	_ runtime.Msg,
) (func(ctx context.Context), error) {
	leftIn, err := io.In.Single("left")
	if err != nil {
		return nil, err
	}

	rightIn, err := io.In.Single("right")
	if err != nil {
		return nil, err
	}

	resOut, err := io.Out.Single("res")
	if err != nil {
		return nil, err
	}

	return func(ctx context.Context) {
		for {
			var leftMsg, rightMsg runtime.Msg
			var leftOk, rightOk bool
			var wg sync.WaitGroup

			wg.Go(func() {
				leftMsg, leftOk = leftIn.Receive(ctx)
			})

			wg.Go(func() {
				rightMsg, rightOk = rightIn.Receive(ctx)
			})

			wg.Wait()

			if !leftOk || !rightOk {
				return
			}

			resMsg := runtime.NewStringMsg(leftMsg.Str() + rightMsg.Str())
			if !resOut.Send(ctx, resMsg) {
				return
			}
		}
	}, nil
}
