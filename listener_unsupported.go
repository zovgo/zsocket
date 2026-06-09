//go:build !windows && !unix

package zsocket

import (
	"context"
	"errors"
	"time"
)

func listen(_ context.Context, _ string, _, _ int32, _, _ bool, _ time.Duration) (Listener, error) {
	return nil, errors.ErrUnsupported
}
