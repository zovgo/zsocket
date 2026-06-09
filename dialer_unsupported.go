//go:build !windows && !unix

package zsocket

import (
	"context"
	"errors"
)

func dial(_ context.Context, _ string) (Conn, error) {
	return nil, errors.ErrUnsupported
}
