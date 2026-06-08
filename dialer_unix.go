//go:build unix

package sock

import (
	"context"
	"net"
	"sync"
)

func dial(parent context.Context, path string) (*pipeConn, error) {
	c, err := (&net.Dialer{}).DialContext(parent, "unix", path)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(parent)
	if d, ok := parent.Deadline(); ok {
		_ = c.SetDeadline(d)
	}
	conn := &pipeConn{Conn: c, ctx: ctx, cancel: cancel}
	conn.wg.Go(conn.catchContext)
	return conn, nil
}

var _ Conn = (*pipeConn)(nil)

type pipeConn struct {
	net.Conn

	ctx    context.Context
	cancel context.CancelFunc

	o  sync.Once
	wg sync.WaitGroup
}

func (c *pipeConn) Read(p []byte) (int, error) {
	if err := c.ctx.Err(); err != nil {
		return 0, err
	}
	return c.Conn.Read(p)
}

func (c *pipeConn) Write(p []byte) (int, error) {
	if err := c.ctx.Err(); err != nil {
		return 0, err
	}
	return c.Conn.Write(p)
}

func (c *pipeConn) Close() (err error) {
	c.o.Do(func() {
		c.cancel()
		err = c.Conn.Close()
	})
	c.wg.Wait()
	return err
}

func (c *pipeConn) catchContext() {
	<-c.ctx.Done()
	c.o.Do(func() {
		_ = c.Conn.Close()
	})
}
