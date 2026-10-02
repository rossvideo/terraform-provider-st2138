package client

import (
	"context"
	"time"

	"github.com/cenkalti/backoff/v5"
	st2138pb "github.com/rossvideo/terraform-provider-st2138/internal/genproto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type retryPolicy struct {
	InitialInterval time.Duration
	MaxInterval     time.Duration
	MaxElapsedTime  time.Duration
	MaxTries        uint
}

var (
	connectPolicy = retryPolicy{
		InitialInterval: 250 * time.Millisecond,
		MaxInterval:     10 * time.Second,
		MaxElapsedTime:  2 * time.Minute,
	}
	setValuePolicy = retryPolicy{
		InitialInterval: 250 * time.Millisecond,
		MaxInterval:     10 * time.Second,
		MaxElapsedTime:  30 * time.Second,
		MaxTries:        6,
	}
)

func (p retryPolicy) options() []backoff.RetryOption {
	b := backoff.NewExponentialBackOff()
	b.InitialInterval = p.InitialInterval
	b.MaxInterval = p.MaxInterval
	return []backoff.RetryOption{
		backoff.WithBackOff(b),
		backoff.WithMaxElapsedTime(p.MaxElapsedTime),
		backoff.WithMaxTries(p.MaxTries),
	}
}

func retry[T any](ctx context.Context, op backoff.Operation[T], p retryPolicy) (T, error) {
	return backoff.Retry(ctx, op, p.options()...)
}

// Overridable in tests.
var (
	retryDial     = retry[*grpc.ClientConn]
	retrySetValue = retry[struct{}]
)

// isRetryableCode reports whether a gRPC status code indicates a transient failure.
func isRetryableCode(c codes.Code) bool {
	switch c {
	case codes.Unavailable, codes.ResourceExhausted, codes.Aborted, codes.DeadlineExceeded:
		return true
	default:
		return false
	}
}

// setValueRetryInterceptor retries SetValue RPCs with exponential backoff on transient errors.
func setValueRetryInterceptor(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	if method != st2138pb.CatenaService_SetValue_FullMethodName {
		return invoker(ctx, method, req, reply, cc, opts...)
	}
	_, err := retrySetValue(ctx, func() (struct{}, error) {
		err := invoker(ctx, method, req, reply, cc, opts...)
		if err != nil && !isRetryableCode(status.Code(err)) {
			return struct{}{}, backoff.Permanent(err)
		}
		return struct{}{}, err
	}, setValuePolicy)
	return err
}
