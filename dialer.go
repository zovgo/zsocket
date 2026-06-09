package zsocket

import (
	"context"
	"net"
	"runtime"
	"time"
)

type Conn interface {
	net.Conn
}

type Dialer struct {
	PrependWindowsPrefix bool
}

func (d Dialer) Dial(path string) (Conn, error) {
	return d.DialTimeout(path, time.Second*10)
}

func (d Dialer) DialTimeout(path string, dur time.Duration) (Conn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dur)
	defer cancel()
	return d.DialContext(ctx, path)
}

func (d Dialer) DialContext(ctx context.Context, path string) (Conn, error) {
	if d.PrependWindowsPrefix {
		path = WindowsSocketPrefix + path
	}
	return dial(ctx, path)
}

var DefaultDialer = Dialer{PrependWindowsPrefix: runtime.GOOS == "windows"}

func Dial(path string) (Conn, error) {
	return DialTimeout(path, time.Second*10)
}

func DialTimeout(path string, timeout time.Duration) (Conn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return DialContext(ctx, path)
}

func DialContext(ctx context.Context, path string) (Conn, error) {
	return DefaultDialer.DialContext(ctx, path)
}
