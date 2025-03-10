package funcs

import (
	"context"
	"strings"

	"github.com/nevalang/neva/internal/runtime"
)

type stringsSplit struct{}

func (p stringsSplit) Create(io runtime.IO, _ runtime.Msg) (func(ctx context.Context), error) {
	dataIn, err := io.In.Single("data")
	if err != nil {
		return nil, err
	}

	delimIn, err := io.In.Single("delim")
	if err != nil {
		return nil, err
	}

	resOut, err := io.Out.Single("res")
	if err != nil {
		return nil, err
	}

	return func(ctx context.Context) {
		for {
			data, ok := dataIn.Receive(ctx)
			if !ok {
				return
			}

			delim, ok := delimIn.Receive(ctx)
			if !ok {
				return
			}

			splitted := strings.Split(data.Str(), delim.Str())
			res := make([]runtime.Msg, len(splitted))
			for i, s := range splitted {
				res[i] = runtime.NewStringMsg(s)
			}

			if !resOut.Send(ctx, runtime.NewListMsg(res)) {
				return
			}
		}
	}, nil
}
