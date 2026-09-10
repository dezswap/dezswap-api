package dashboard

import (
	"testing"

	ds "github.com/dezswap/dezswap-api/api/v1/service/dashboard"
	"github.com/stretchr/testify/assert"
)

func TestParseTokenAddrs(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected []ds.Addr
	}{
		{
			name:     "Empty String",
			input:    "",
			expected: []ds.Addr(nil),
		},
		{
			name:     "Empty Tokens",
			input:    ", ",
			expected: []ds.Addr(nil),
		},
		{
			name:     "A Single Token",
			input:    "axpla",
			expected: []ds.Addr{"axpla"},
		},
		{
			name:     "Multiple Tokens",
			input:    "axpla,xpla1abcd,ibc/ABCD1234",
			expected: []ds.Addr{"axpla", "xpla1abcd", "ibc/ABCD1234"},
		},
		{
			name:     "Multiple Tokens with Whitespace",
			input:    "axpla, xpla1abcd ,ibc/ABCD1234",
			expected: []ds.Addr{"axpla", "xpla1abcd", "ibc/ABCD1234"},
		},
		{
			name:     "Multiple Tokens Including Empty One",
			input:    "axpla,,ibc/ABCD1234",
			expected: []ds.Addr{"axpla", "ibc/ABCD1234"},
		},
		{
			name:     "Multiple Tokens Including Whitespace Token",
			input:    "axpla,  ,ibc/ABCD1234",
			expected: []ds.Addr{"axpla", "ibc/ABCD1234"},
		},
		{
			name:     "Starts with Whitespace",
			input:    " axpla,  ,ibc/ABCD1234",
			expected: []ds.Addr{"axpla", "ibc/ABCD1234"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := parseTokenAddrs(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
