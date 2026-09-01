package repo

import (
	"strings"
	"testing"

	"github.com/dezswap/dezswap-api/indexer"
	"github.com/dezswap/dezswap-api/pkg"
	"github.com/dezswap/dezswap-api/pkg/types"
	xpla_mock "github.com/dezswap/dezswap-api/pkg/xpla/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type assetRepoSuite struct {
	suite.Suite
	client          pkg.Client
	r               assetRepoImpl
	networkMetadata pkg.NetworkMetadata
}

func (s *assetRepoSuite) SetupSuite() {
	s.client = xpla_mock.NewClientMock()
	s.networkMetadata = pkg.NewNetworkMetadata(pkg.NetworkMetadataConfig{
		NetworkName:   pkg.NetworkNameXplaChain,
		AddrPrefix:    "xpla1",
		TokenPrefixes: map[types.TokenType]string{types.TokenTypeCW20: "xcw20:", types.TokenTypeERC20: "xerc20:"},
		BlockSecond:   5,
		Mainnet:       pkg.ChainInfo{ChainIdPrefix: "dimension"},
		Testnets:      []pkg.ChainInfo{{ChainIdPrefix: "cube"}},
	})
	s.r = assetRepoImpl{s.client, &assetMapperImpl{}, s.networkMetadata}
}

func (s *assetRepoSuite) Test_VerifiedTokens() {
	tcs := []struct {
		chainID              string
		expected             []indexer.Token
		verifiedCw20sResult  *types.TokensRes
		verifiedIbcsResult   *types.IbcsRes
		verifiedErc20sResult *types.TokensRes
		err                  error
	}{
		{
			chainID: "cube_47-5",
			expected: []indexer.Token{
				{
					Address:  "xerc20:217c395CDC38D55d1F83528df05b9412cde5b800",
					ChainId:  "cube_47-5",
					Protocol: "XPLA",
					Symbol:   "ZAD",
					Name:     "ZAD",
					Decimals: 18,
					Icon:     "https://assets.xpla.io/icon/evm/ZAD.png",
					Verified: true,
				},
			},
			verifiedCw20sResult: &types.TokensRes{
				Mainnet: types.TokenResMap{}, Testnet: types.TokenResMap{},
			},
			verifiedIbcsResult: &types.IbcsRes{
				Mainnet: types.IbcResMap{}, Testnet: types.IbcResMap{},
			},
			verifiedErc20sResult: &types.TokensRes{
				Mainnet: types.TokenResMap{
					"0x26D086423f64d339481f379F8988004B4fcaB093": types.TokenRes{
						Protocol: strPtr("XPLA"),
						Symbol:   strPtr("NINKY"),
						Name:     strPtr("Idle Ninja Online Token"),
						Token:    strPtr("0x26D086423f64d339481f379F8988004B4fcaB093"),
						Icon:     strPtr("https://assets.xpla.io/icon/evm/xNINKY.png"),
						Decimals: u8Ptr(18),
					},
				}, Testnet: types.TokenResMap{
					"0x217c395CDC38D55d1F83528df05b9412cde5b800": types.TokenRes{
						Protocol: strPtr("XPLA"),
						Symbol:   strPtr("ZAD"),
						Name:     strPtr("ZAD"),
						Token:    strPtr("0x217c395CDC38D55d1F83528df05b9412cde5b800"),
						Icon:     strPtr("https://assets.xpla.io/icon/evm/ZAD.png"),
						Decimals: u8Ptr(18),
					},
				},
			},
			err: nil,
		},
		{
			chainID:              "cube_47-5",
			expected:             []indexer.Token{},
			verifiedCw20sResult:  &types.TokensRes{},
			verifiedIbcsResult:   &types.IbcsRes{},
			verifiedErc20sResult: &types.TokensRes{},
			err:                  nil,
		},
	}

	for _, tc := range tcs {
		// prepare mock response
		{
			s.client.(*xpla_mock.ClientMock).On("VerifiedCw20s").Return(tc.verifiedCw20sResult, tc.err).Once()
			s.client.(*xpla_mock.ClientMock).On("VerifiedIbcs").Return(tc.verifiedIbcsResult, tc.err).Once()
			s.client.(*xpla_mock.ClientMock).On("VerifiedErc20s").Return(tc.verifiedErc20sResult, tc.err).Once()
		}

		actual, err := s.r.VerifiedTokens(tc.chainID)
		if err != nil {
			assert.True(s.T(), strings.Contains(err.Error(), tc.err.Error()))
		} else {
			assert.Equal(s.T(), tc.expected, actual)
		}
	}
}

func Test_AssetRepo(t *testing.T) {
	suite.Run(t, new(assetRepoSuite))
}

func Test_NewAssetRepo(t *testing.T) {
	networkMetadata := pkg.NewNetworkMetadata(pkg.NetworkMetadataConfig{
		NetworkName:   pkg.NetworkNameXplaChain,
		AddrPrefix:    "xpla1",
		TokenPrefixes: map[types.TokenType]string{types.TokenTypeCW20: "xcw20:", types.TokenTypeERC20: "xerc20:"},
		BlockSecond:   5,
		Mainnet:       pkg.ChainInfo{ChainIdPrefix: "dimension", FactoryAddress: "xpla1abcd"},
		Testnets:      []pkg.ChainInfo{{ChainIdPrefix: "cube", FactoryAddress: "xpla1efgh"}},
	})

	t.Run("success with valid factory address", func(t *testing.T) {
		repo, err := NewAssetRepo(networkMetadata, "cube_47-5", "xpla1efgh")
		assert.NoError(t, err)
		assert.NotNil(t, repo)
	})

	// every supported network must resolve to a client, otherwise the repo panics on first use
	t.Run("client is set for every supported network", func(t *testing.T) {
		asiAllianceMetadata := pkg.NewNetworkMetadata(pkg.NetworkMetadataConfig{
			NetworkName:   pkg.NetworkNameAsiAlliance,
			AddrPrefix:    "fetch1",
			TokenPrefixes: map[types.TokenType]string{},
			BlockSecond:   5,
			Mainnet:       pkg.ChainInfo{ChainIdPrefix: "fetchhub", FactoryAddress: "fetch1abcd"},
			Testnets:      []pkg.ChainInfo{{ChainIdPrefix: "dorado", FactoryAddress: "fetch1efgh"}},
		})
		terraClassicMetadata := pkg.NewNetworkMetadata(pkg.NetworkMetadataConfig{
			NetworkName:   pkg.NetworkNameTerraClassic,
			AddrPrefix:    "terra1",
			TokenPrefixes: map[types.TokenType]string{},
			BlockSecond:   6,
			Mainnet:       pkg.ChainInfo{ChainIdPrefix: "columbus", FactoryAddress: "terra1abcd"},
		})

		tcs := []struct {
			name     string
			metadata pkg.NetworkMetadata
			chainId  string
			factory  string
		}{
			{"xpla chain", networkMetadata, "cube_47-5", "xpla1efgh"},
			{"asi alliance mainnet", asiAllianceMetadata, "fetchhub-4", "fetch1abcd"},
			{"asi alliance testnet", asiAllianceMetadata, "dorado-1", "fetch1efgh"},
			{"terra classic", terraClassicMetadata, "columbus-5", "terra1abcd"},
		}

		for _, tc := range tcs {
			t.Run(tc.name, func(t *testing.T) {
				repo, err := NewAssetRepo(tc.metadata, tc.chainId, tc.factory)
				assert.NoError(t, err)

				impl, ok := repo.(*assetRepoImpl)
				require.Truef(t, ok, "expected *assetRepoImpl but got %T", repo)
				assert.NotNil(t, impl.Client)
			})
		}
	})

	t.Run("error with unregistered factory address", func(t *testing.T) {
		repo, err := NewAssetRepo(networkMetadata, "cube_47-5", "invalid_factory_address")
		assert.Error(t, err)
		assert.Equal(t, pkg.ErrUnregisteredFactoryAddress, err)
		assert.Nil(t, repo)
	})

	t.Run("error with unsupported network", func(t *testing.T) {
		unsupportedMetadata := pkg.NewNetworkMetadata(pkg.NetworkMetadataConfig{
			NetworkName:   "unsupported",
			AddrPrefix:    "uns1",
			TokenPrefixes: map[types.TokenType]string{},
			BlockSecond:   5,
			Mainnet:       pkg.ChainInfo{ChainIdPrefix: "mainprefix", FactoryAddress: "mainfactory"},
			Testnets:      []pkg.ChainInfo{{ChainIdPrefix: "testprefix", FactoryAddress: "testfactory"}},
		})
		repo, err := NewAssetRepo(unsupportedMetadata, "testprefix", "testfactory")
		assert.Error(t, err)
		assert.Equal(t, pkg.ErrUnsupportedNetwork, err)
		assert.Nil(t, repo)
	})
}

func strPtr(s string) *string { return &s }
func u8Ptr(i uint8) *uint8    { return &i }
