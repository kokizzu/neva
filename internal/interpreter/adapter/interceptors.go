package adapter

import (
	"fmt"
	"os"

	"github.com/nevalang/neva/internal/runtime"
)

type debugInterceptor struct {
	logger Logger
}

type Logger interface {
	Printf(format string, v ...any) error
}

type stdLogger struct{}

func (s stdLogger) Printf(format string, v ...any) error {
	_, err := fmt.Printf(format, v...)
	return err
}

type fileLogger struct {
	filepath string
}

func (f fileLogger) Printf(format string, v ...any) error {
	file, err := os.OpenFile(f.filepath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = fmt.Fprintf(file, format, v...)
	return err
}

func (d debugInterceptor) Sent(sender runtime.PortSlotAddr, msg runtime.Msg) runtime.Msg {
	d.logger.Printf("sent from: %v, msg: %v\n", d.formatPortSlotAddr(sender), d.formatMsg(msg))
	return msg
}

func (d debugInterceptor) Received(receiver runtime.PortSlotAddr, msg runtime.Msg) runtime.Msg {
	d.logger.Printf("received to: %v, msg: %v\n", d.formatPortSlotAddr(receiver), d.formatMsg(msg))
	return msg
}

func (d debugInterceptor) formatMsg(msg runtime.Msg) string {
	if s, ok := msg.(runtime.StrMsg); ok {
		return fmt.Sprintf(`"%s"`, s)
	}
	return fmt.Sprint(msg)
}

func (d debugInterceptor) formatPortSlotAddr(slotAddr runtime.PortSlotAddr) string {
	s := fmt.Sprintf("%v:%v", slotAddr.Path, slotAddr.Port)
	if slotAddr.Index != nil {
		s = fmt.Sprintf("%v[%v]", s, *slotAddr.Index)
	}
	return s
}

type prodInterceptor struct{}

func (prodInterceptor) Sent(sender runtime.PortSlotAddr, msg runtime.Msg) runtime.Msg {
	return msg
}

func (prodInterceptor) Received(receiver runtime.PortSlotAddr, msg runtime.Msg) runtime.Msg {
	return msg
}
