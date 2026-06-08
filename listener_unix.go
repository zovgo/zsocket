//go:build unix

package sock

import (
	"context"
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

func listen(parent context.Context, path string, _, _ int32, _, errorIfUsed bool, timeout time.Duration) (*pipeListener, error) {
	if err := parent.Err(); err != nil {
		return nil, err
	}
	if errorIfUsed && tryDial(parent, path, timeout) {
		return nil, ErrSocketUsedByAnotherProcess
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("remove prev file: %w", err)
	}
	li, err := net.Listen("unix", path)
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

func tryDial(parent context.Context, path string, timeout time.Duration) bool {
	ctx, cancel := context.WithTimeout(parent, timeout)
	conn, err := dial(ctx, path)
	cancel()
	if err == nil {
		_ = conn.Close()
		return true
	}
	return false
}

var _ Listener = (*pipeListener)(nil)

type pipeListener struct {
	addr string
	li   net.Listener

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
		_ = os.Remove(l.addr)
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
		_ = os.Remove(l.addr)
	})
}
