package pkg

import (
	"net"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Test_NewGrpcClientQueryTimeout(t *testing.T) {
	tcs := []struct {
		name     string
		timeout  time.Duration
		expected time.Duration
	}{
		{"configured timeout is kept", 15 * time.Second, 15 * time.Second},
		{"zero falls back to the default", 0, NodeQueryTimeout},
		{"negative falls back to the default", -time.Second, NodeQueryTimeout},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			c, err := NewGrpcClientWithTimeout("localhost:9090", false, tc.timeout)
			require.NoError(t, err)

			impl, ok := c.(*grpcClient)
			require.True(t, ok)
			t.Cleanup(func() { _ = impl.Close() })

			require.Equal(t, tc.expected, impl.queryTimeout)
		})
	}
}

func Test_NewGrpcClientUsesDefaultTimeout(t *testing.T) {
	c, err := NewGrpcClient("localhost:9090", false)
	require.NoError(t, err)

	impl, ok := c.(*grpcClient)
	require.True(t, ok)
	t.Cleanup(func() { _ = impl.Close() })

	require.Equal(t, NodeQueryTimeout, impl.queryTimeout)
}

// startStalledServer starts a gRPC server that never answers any RPC, so a call
// can only end at the client side deadline.
func startStalledServer(t *testing.T) string {
	t.Helper()

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	srv := grpc.NewServer(grpc.UnknownServiceHandler(func(_ any, stream grpc.ServerStream) error {
		<-stream.Context().Done()
		return stream.Context().Err()
	}))
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.Stop)

	return lis.Addr().String()
}

func Test_GrpcClientQueriesRespectTimeout(t *testing.T) {
	target := startStalledServer(t)
	const timeout = 200 * time.Millisecond

	tcs := []struct {
		name string
		call func(GrpcClient) error
	}{
		{"SyncedHeight", func(c GrpcClient) error { _, err := c.SyncedHeight(); return err }},
		{"QueryContract", func(c GrpcClient) error { _, err := c.QueryContract("addr", []byte("{}"), 0); return err }},
		{"QueryContract with height", func(c GrpcClient) error { _, err := c.QueryContract("addr", []byte("{}"), 42); return err }},
		{"QueryIbcDenomTrace", func(c GrpcClient) error { _, err := c.QueryIbcDenomTrace("hash"); return err }},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			c, err := NewGrpcClientWithTimeout(target, false, timeout)
			require.NoError(t, err)
			t.Cleanup(func() { _ = c.(*grpcClient).Close() })

			start := time.Now()
			done := make(chan error, 1)
			go func() { done <- tc.call(c) }()

			select {
			case err := <-done:
				require.Error(t, err)
				require.Equal(t, codes.DeadlineExceeded, status.Code(errors.Cause(err)))
				require.GreaterOrEqual(t, time.Since(start), timeout)
			case <-time.After(10 * timeout):
				t.Fatal("query did not honour the configured timeout")
			}
		})
	}
}
