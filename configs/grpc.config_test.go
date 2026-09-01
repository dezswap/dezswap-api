package configs

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGrpcConfigOverrideQueryTimeoutSec(t *testing.T) {
	tcs := []struct {
		name     string
		lhs      GrpcConfig
		rhs      GrpcConfig
		expected int
	}{
		{
			name:     "a set value overrides",
			lhs:      GrpcConfig{QueryTimeoutSec: 5},
			rhs:      GrpcConfig{QueryTimeoutSec: 15},
			expected: 15,
		},
		{
			name:     "an unset value keeps the original",
			lhs:      GrpcConfig{QueryTimeoutSec: 5},
			rhs:      GrpcConfig{Host: "env.example.com"},
			expected: 5,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.lhs.Override(tc.rhs)
			require.Equal(t, tc.expected, tc.lhs.QueryTimeoutSec)
		})
	}
}

func TestGrpcConfigIsZero(t *testing.T) {
	require.True(t, GrpcConfig{}.IsZero())
	require.False(t, GrpcConfig{Host: "a.example.com"}.IsZero())
	require.False(t, GrpcConfig{Port: "443"}.IsZero())
	require.False(t, GrpcConfig{UseTls: true}.IsZero())
	// A lone query_timeout_sec does not name a node. Callers read a non-zero config as
	// "the node was overridden" and drop the src_nodes list, so counting a timeout
	// here would silently disable the fallback candidates.
	require.True(t, GrpcConfig{QueryTimeoutSec: 1}.IsZero())
}
