package pkg

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/types"
	"strings"
)

func NewDecFromStrWithTruncate(input string) (types.Dec, error) {
	truncatedInput := truncateDecimal(input)
	dec, err := types.NewDecFromStr(truncatedInput)
	if err != nil {
		return types.Dec{}, fmt.Errorf("failed to parse decimal: %w", err)
	}

	return dec, nil
}

func truncateDecimal(input string) string {
	parts := strings.Split(input, ".")
	if len(parts) == 1 {
		return input
	}

	fractional := parts[1]
	if len(fractional) > types.Precision {
		fractional = fractional[:types.Precision]
	}

	return parts[0] + "." + fractional
}
