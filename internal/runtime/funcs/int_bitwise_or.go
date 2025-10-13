package funcs

import (
	"context"
	"sync"

	"github.com/nevalang/neva/internal/runtime"
)

type intBitwiseOr struct{}

func (intBitwiseOr) Create(
	io runtime.IO,
	_ runtime.Msg,
) (func(ctx context.Context), error) {
	accIn, err := io.In.Single("left")
	if err != nil {
		return nil, err
	}

	elIn, err := io.In.Single("right")
	if err != nil {
		return nil, err
	}

	resOut, err := io.Out.Single("res")
	if err != nil {
		return nil, err
	}

	return func(ctx context.Context) {
		for {
			var accMsg, elMsg runtime.Msg
			var accOk, elOk bool

			var wg sync.WaitGroup

			wg.Go(func() {
				accMsg, accOk = accIn.Receive(ctx)
			})

			wg.Go(func() {
				elMsg, elOk = elIn.Receive(ctx)
			})

			wg.Wait()

			if !accOk || !elOk {
				return
			}

			if !resOut.Send(ctx, runtime.NewIntMsg(accMsg.Int()|elMsg.Int())) {
				return
			}
		}
	}, nil
}
