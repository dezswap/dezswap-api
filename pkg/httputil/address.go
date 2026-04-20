package httputil

import "strings"

// DecodeAddressParam restores IBC token addresses that were encoded for
// URL path segments. by replacing a leading "ibc-" prefix with "ibc/".
func DecodeAddressParam(encoded string) string {
	if strings.HasPrefix(encoded, "ibc-") {
		return "ibc/" + encoded[len("ibc-"):]
	}
	return encoded
}
