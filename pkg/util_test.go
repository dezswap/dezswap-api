package pkg

import (
	"testing"
)

func Test_TruncateDecimal(t *testing.T) {
	tcs := []struct {
		input    string
		expected string
	}{
		{"0.0035453499282647604", "0.003545349928264760"},
		{"1.1234567890123456789", "1.123456789012345678"},
		{"123.4567890123456789", "123.4567890123456789"},
		{"123456", "123456"},
		{"0.0000000000000000001", "0.000000000000000000"},
		{"-1.0000000000000000009", "-1.000000000000000000"},
	}

	for _, tc := range tcs {
		t.Run(tc.input, func(t *testing.T) {
			result := truncateDecimal(tc.input)
			if result != tc.expected {
				t.Errorf("expected %s but got %s", tc.expected, result)
			}
		})
	}
}

func TestSuite(t *testing.T) {
	t.Run("TestTruncateDecimal", Test_TruncateDecimal)
}
