package job

import (
	"context"
	"time"
)

type LockStore interface {
	TryAcquire(ctx context.Context, name, owner string, timeout time.Duration) (acquired bool, err error)
	Release(ctx context.Context, name, owner string) error
	ReleaseExpired(ctx context.Context) (int64, error)
}
