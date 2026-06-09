package zsocket

import (
	"context"
	"errors"
	"io"
	"net"
	"time"
)

type Listener interface {
	io.Closer
	Accept() (Conn, error)
	Addr() string
	Original() net.Listener
}

const (
	defaultInputBuffer  = 1024 * 512
	defaultOutputBuffer = 1024 * 512
)

type ListenConfig struct {
	Context              context.Context
	InputBuffer          int32
	OutputBuffer         int32
	NotMessageMode       bool
	PrependWindowsPrefix bool
	ErrorIfAlreadyUsed   bool
	InitialDialTimeout   time.Duration
}

const WindowsSocketPrefix = `\\.\pipe\`

func (conf ListenConfig) Listen(path string) (Listener, error) {
	if conf.Context == nil {
		conf.Context = context.Background()
	}
	if conf.InputBuffer <= 0 {
		conf.InputBuffer = defaultInputBuffer
	}
	if conf.OutputBuffer <= 0 {
		conf.OutputBuffer = defaultOutputBuffer
	}
	if conf.InitialDialTimeout <= 0 {
		conf.InitialDialTimeout = time.Second * 5
	}
	if conf.PrependWindowsPrefix {
		path = WindowsSocketPrefix + path
	}
	return listen(
		conf.Context,
		path,
		conf.InputBuffer,
		conf.OutputBuffer,
		conf.NotMessageMode,
		conf.ErrorIfAlreadyUsed,
		conf.InitialDialTimeout,
	)
}

var ErrSocketUsedByAnotherProcess = errors.New("socket used by another process")
