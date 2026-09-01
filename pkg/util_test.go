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
			if actual := tc.metadata.IsMainnet(tc.chainId); actual != tc.expectedMainnet {
				t.Errorf("IsMainnet(%q): expected %v but got %v", tc.chainId, tc.expectedMainnet, actual)
			}
			if actual := tc.metadata.IsTestnet(tc.chainId); actual != tc.expectedTestnet {
				t.Errorf("IsTestnet(%q): expected %v but got %v", tc.chainId, tc.expectedTestnet, actual)
			}
			expectedEither := tc.expectedMainnet || tc.expectedTestnet
			if actual := tc.metadata.IsMainnetOrTestnet(tc.chainId); actual != expectedEither {
				t.Errorf("IsMainnetOrTestnet(%q): expected %v but got %v", tc.chainId, expectedEither, actual)
			}

			factory, err := tc.metadata.GetFactoryAddress(tc.chainId)
			if err != tc.expectedErr {
				t.Errorf("GetFactoryAddress(%q): expected error %v but got %v", tc.chainId, tc.expectedErr, err)
			}
			if factory != tc.expectedFactory {
				t.Errorf("GetFactoryAddress(%q): expected %q but got %q", tc.chainId, tc.expectedFactory, factory)
			}
		})
	}
}
