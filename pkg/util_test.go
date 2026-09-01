package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func Test_NetworkMetadataChainClassification(t *testing.T) {
	// a network with more than one testnet
	multiTestnet := NewNetworkMetadata(NetworkMetadataConfig{
		NetworkName: "multi",
		AddrPrefix:  "multi1",
		Mainnet:     ChainInfo{ChainIdPrefix: "mainchain", FactoryAddress: "multi1mainfactory"},
		Testnets: []ChainInfo{
			{ChainIdPrefix: "testchain", FactoryAddress: "multi1testfactory"},
			{ChainIdPrefix: "devchain", FactoryAddress: "multi1devfactory"},
		},
	})
	// a network without any testnet, as terra classic
	mainnetOnly := NewNetworkMetadata(NetworkMetadataConfig{
		NetworkName: "mainnet-only",
		AddrPrefix:  "only1",
		Mainnet:     ChainInfo{ChainIdPrefix: "mainchain", FactoryAddress: "only1mainfactory"},
	})
	// chains without a prefix cannot be matched, so they must stay unregistered
	prefixless := NewNetworkMetadata(NetworkMetadataConfig{
		NetworkName: "prefixless",
		AddrPrefix:  "none1",
		Mainnet:     ChainInfo{FactoryAddress: "none1mainfactory"},
		Testnets:    []ChainInfo{{FactoryAddress: "none1testfactory"}},
	})

	tcs := []struct {
		name            string
		metadata        NetworkMetadata
		chainId         string
		expectedMainnet bool
		expectedTestnet bool
		expectedFactory string
		expectedErr     error
	}{
		{"multi testnet: mainnet", multiTestnet, "mainchain-1", true, false, "multi1mainfactory", nil},
		{"multi testnet: first testnet", multiTestnet, "testchain-1", false, true, "multi1testfactory", nil},
		{"multi testnet: second testnet", multiTestnet, "devchain-2", false, true, "multi1devfactory", nil},
		{"multi testnet: unregistered chain", multiTestnet, "otherchain-1", false, false, "", ErrUnsupportedNetwork},
		// a prefix must match at the start of the chain id, not anywhere in it
		{"multi testnet: prefix not at the start", multiTestnet, "notmainchain-1", false, false, "", ErrUnsupportedNetwork},
		{"mainnet only: mainnet", mainnetOnly, "mainchain-1", true, false, "only1mainfactory", nil},
		{"mainnet only: unregistered chain", mainnetOnly, "testchain-1", false, false, "", ErrUnsupportedNetwork},
		{"prefixless: any chain", prefixless, "mainchain-1", false, false, "", ErrUnsupportedNetwork},
		{"prefixless: empty chain id", prefixless, "", false, false, "", ErrUnsupportedNetwork},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedMainnet, tc.metadata.IsMainnet(tc.chainId), "IsMainnet(%q)", tc.chainId)
			assert.Equal(t, tc.expectedTestnet, tc.metadata.IsTestnet(tc.chainId), "IsTestnet(%q)", tc.chainId)
			expectedEither := tc.expectedMainnet || tc.expectedTestnet
			assert.Equal(t, expectedEither, tc.metadata.IsMainnetOrTestnet(tc.chainId), "IsMainnetOrTestnet(%q)", tc.chainId)

			factory, err := tc.metadata.GetFactoryAddress(tc.chainId)
			assert.Equal(t, tc.expectedErr, err, "GetFactoryAddress(%q)", tc.chainId)
			assert.Equal(t, tc.expectedFactory, factory, "GetFactoryAddress(%q)", tc.chainId)
		})
	}
}

func Test_GetNetworkMetadata(t *testing.T) {
	tcs := []struct {
		name            string
		chainId         string
		expectedName    NetworkName
		expectedMainnet bool
		expectedTestnet bool
		expectedErr     error
	}{
		{"xpla mainnet", "dimension_37-1", NetworkNameXplaChain, true, false, nil},
		{"xpla testnet", "cube_47-5", NetworkNameXplaChain, false, true, nil},
		{"asi alliance mainnet", "fetchhub-4", NetworkNameAsiAlliance, true, false, nil},
		{"asi alliance testnet", "dorado-1", NetworkNameAsiAlliance, false, true, nil},
		{"terra classic mainnet", "columbus-5", NetworkNameTerraClassic, true, false, nil},
		// terra classic has no testnet, so no unregistered chain may resolve to it
		{"terra 2 is not registered", "phoenix-1", "", false, false, ErrUnsupportedNetwork},
		{"unknown chain", "unknown-1", "", false, false, ErrUnsupportedNetwork},
		{"empty chain id", "", "", false, false, ErrUnsupportedNetwork},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			metadata, err := GetNetworkMetadata(tc.chainId)
			require.Equal(t, tc.expectedErr, err)
			assert.Equal(t, tc.expectedName, metadata.NetworkName)
			if tc.expectedErr != nil {
				return
			}
			assert.Equal(t, tc.expectedMainnet, metadata.IsMainnet(tc.chainId), "IsMainnet")
			assert.Equal(t, tc.expectedTestnet, metadata.IsTestnet(tc.chainId), "IsTestnet")

			factory, err := metadata.GetFactoryAddress(tc.chainId)
			require.NoError(t, err, "GetFactoryAddress")
			assert.NotEmpty(t, factory, "GetFactoryAddress: expected a registered factory address")
		})
	}
}
