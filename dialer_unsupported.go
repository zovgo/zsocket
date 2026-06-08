//go:build !windows && !unix

package sock

import (
	"context"
	"errors"
)

func dial(_ context.Context, _ string) (Conn, error) {
	return nil, errors.ErrUnsupported
}
