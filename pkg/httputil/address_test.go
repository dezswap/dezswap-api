package httputil

import (
	"testing"
)

func TestDecodeAddressParam(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"", ""},
		{"xpla1abc", "xpla1abc"},
		{"ibc-ABCD1234", "ibc/ABCD1234"},
		{"some-other-denom", "some-other-denom"},
	}

	for _, tc := range cases {
		got := DecodeAddressParam(tc.input)
		if got != tc.want {
			t.Errorf("DecodeAddressParam(%q) = %q; want %q", tc.input, got, tc.want)
		}
	}
}
