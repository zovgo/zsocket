//go:build windows

package zsocket

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/Microsoft/go-winio"
)

const secDesc = "D:(A;;GA;;;BA)(A;;GA;;;SY)(A;;GA;;;AU)"

func listen(parent context.Context, path string, inpBuff, outBuff int32, notMsgMode, _ bool, _ time.Duration) (*pipeListener, error) {
	if err := parent.Err(); err != nil {
		return nil, err
	}
	conf := &winio.PipeConfig{
		SecurityDescriptor: secDesc,
		MessageMode:        !notMsgMode,
		InputBufferSize:    inpBuff,
		OutputBufferSize:   outBuff,
	}
	li, err := winio.ListenPipe(path, conf)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(parent)
	pl := &pipeListener{
		addr:   path,
		li:     li,
		ctx:    ctx,
		cancel: cancel,
	}
	pl.wg.Go(pl.catchContext)
	return pl, nil
}

var _ Listener = (*pipeListener)(nil)

type pipeListener struct {
	addr string

	li net.Listener

	o  sync.Once
	wg sync.WaitGroup

	ctx    context.Context
	cancel context.CancelFunc
}

func (l *pipeListener) Accept() (Conn, error) {
	select {
	case <-l.ctx.Done():
		return nil, l.ctx.Err()
	default:
	}
	return l.li.Accept()
}

func (l *pipeListener) Close() (err error) {
	l.o.Do(func() {
		l.cancel()
		err = l.li.Close()
	})
	l.wg.Wait()
	return err
}

func (l *pipeListener) Addr() string {
	return l.addr
}

func (l *pipeListener) Original() net.Listener {
	return l.li
}

func (l *pipeListener) catchContext() {
	<-l.ctx.Done()
	l.o.Do(func() {
		_ = l.li.Close()
	})
}
