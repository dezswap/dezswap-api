package pkg

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
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
