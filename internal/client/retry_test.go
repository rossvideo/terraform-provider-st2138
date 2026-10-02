package client

import (
	"context"
	"errors"
	"testing"

	"github.com/cenkalti/backoff/v5"
	st2138pb "github.com/rossvideo/terraform-provider-st2138/internal/genproto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestEnsureConn_UsesConnectPolicy(t *testing.T) {
	orig := retryDial
	t.Cleanup(func() { retryDial = orig })

	sentinel := errors.New("stub")
	var gotPolicy retryPolicy
	called := false
	retryDial = func(ctx context.Context, op backoff.Operation[*grpc.ClientConn], p retryPolicy) (*grpc.ClientConn, error) {
		called = true
		gotPolicy = p
		return nil, sentinel
	}

	c := &Client{Transport: "grpc", Endpoint: "localhost:6254"}
	if err := c.ensureConn(context.Background()); !errors.Is(err, sentinel) {
		t.Fatalf("ensureConn() error = %v, want %v", err, sentinel)
	}
	if !called {
		t.Fatal("retryDial was not called")
	}
	if gotPolicy != connectPolicy {
		t.Errorf("policy = %+v, want %+v", gotPolicy, connectPolicy)
	}
}

func TestSetValueRetryInterceptor_UsesSetValuePolicy(t *testing.T) {
	orig := retrySetValue
	t.Cleanup(func() { retrySetValue = orig })

	tests := []struct {
		name          string
		invokeErr     error
		wantPermanent bool
	}{
		{"success", nil, false},
		{"unavailable is retryable", status.Error(codes.Unavailable, "down"), false},
		{"deadline is retryable", status.Error(codes.DeadlineExceeded, "slow"), false},
		{"invalid argument is permanent", status.Error(codes.InvalidArgument, "bad"), true},
		{"plain error is permanent", errors.New("boom"), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotPolicy retryPolicy
			var opErr error
			retrySetValue = func(ctx context.Context, op backoff.Operation[struct{}], p retryPolicy) (struct{}, error) {
				gotPolicy = p
				_, opErr = op()
				return struct{}{}, opErr
			}
			invoker := func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
				return tt.invokeErr
			}

			_ = setValueRetryInterceptor(context.Background(), st2138pb.CatenaService_SetValue_FullMethodName, nil, nil, nil, invoker)

			if gotPolicy != setValuePolicy {
				t.Errorf("policy = %+v, want %+v", gotPolicy, setValuePolicy)
			}
			var perm *backoff.PermanentError
			if isPerm := errors.As(opErr, &perm); isPerm != tt.wantPermanent {
				t.Errorf("permanent = %v, want %v (err %v)", isPerm, tt.wantPermanent, opErr)
			}
			if tt.invokeErr == nil && opErr != nil {
				t.Errorf("op error = %v, want nil", opErr)
			}
		})
	}
}

func TestSetValueRetryInterceptor_SkipsOtherMethods(t *testing.T) {
	orig := retrySetValue
	t.Cleanup(func() { retrySetValue = orig })

	retrySetValue = func(ctx context.Context, op backoff.Operation[struct{}], p retryPolicy) (struct{}, error) {
		t.Fatal("retrySetValue should not be called for non-SetValue methods")
		return struct{}{}, nil
	}
	invoked := false
	invoker := func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		invoked = true
		return nil
	}

	if err := setValueRetryInterceptor(context.Background(), st2138pb.CatenaService_GetValue_FullMethodName, nil, nil, nil, invoker); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !invoked {
		t.Error("invoker was not called")
	}
}
