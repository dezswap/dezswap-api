package configs

import (
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestIndexerConfigSrcNodeBackwardCompatibility(t *testing.T) {
	v := viper.New()
	v.Set("indexer.chain_id", "dorado-1")
	v.Set("indexer.src_node.host", "primary.example.com")
	v.Set("indexer.src_node.port", "443")
	v.Set("indexer.src_node.use_tls", true)

	c := indexerConfig(v)

	require.Equal(t, "primary.example.com", c.SrcNode.Host)
	require.Equal(t, "443", c.SrcNode.Port)
	require.True(t, c.SrcNode.UseTls)
	require.Empty(t, c.SrcNodes)
}

func TestIndexerConfigSrcNodes(t *testing.T) {
	v := viper.New()
	v.Set("indexer.chain_id", "dorado-1")
	v.Set("indexer.src_node.host", "primary.example.com")
	v.Set("indexer.src_node.port", "443")
	v.Set("indexer.src_node.use_tls", true)
	v.Set("indexer.src_nodes", []map[string]interface{}{
		{"host": "candidate-1.example.com", "port": "443", "use_tls": true},
		{"host": "candidate-2.example.com", "port": "9090", "use_tls": false},
	})

	c := indexerConfig(v)

	require.Len(t, c.SrcNodes, 2)
	require.Equal(t, GrpcConfig{Host: "candidate-1.example.com", Port: "443", UseTls: true}, c.SrcNodes[0])
	require.Equal(t, GrpcConfig{Host: "candidate-2.example.com", Port: "9090", UseTls: false}, c.SrcNodes[1])
}

func TestIndexerConfigSrcNodesOverrideByEnv(t *testing.T) {
	const envKey = "APP_INDEXER_SRC_NODES"
	require.NoError(t, os.Setenv(envKey, `[{"host":"env-1.example.com","port":"443","use_tls":true},{"host":"env-2.example.com","port":"9090","use_tls":false}]`))
	defer os.Unsetenv(envKey)

	v := newTestViper(t, `
indexer:
  chain_id: dorado-1
  src_nodes:
    - host: candidate-1.example.com
      port: "443"
      use_tls: true
`)

	c := indexerConfig(v)

	require.Equal(t, []GrpcConfig{
		{Host: "env-1.example.com", Port: "443", UseTls: true},
		{Host: "env-2.example.com", Port: "9090", UseTls: false},
	}, c.SrcNodes)
}

func TestIndexerConfigEnvSrcNodeOverrideDisablesSrcNodes(t *testing.T) {
	v := viper.New()
	v.Set("indexer.chain_id", "dorado-1")
	v.Set("indexer.src_node.host", "primary.example.com")
	v.Set("indexer.src_node.port", "443")
	v.Set("indexer.src_node.use_tls", true)
	v.Set("indexer.src_nodes", []map[string]interface{}{
		{"host": "candidate-1.example.com", "port": "443", "use_tls": true},
	})
	v.Set("INDEXER_SRC_NODE_HOST", "env.example.com")
	v.Set("INDEXER_SRC_NODE_PORT", "8443")

	c := indexerConfig(v)

	require.Equal(t, "env.example.com", c.SrcNode.Host)
	require.Equal(t, "8443", c.SrcNode.Port)
	require.True(t, c.SrcNode.UseTls)
	require.Empty(t, c.SrcNodes)
}

func TestIndexerConfigSrcNodesEnvUnmarshalErrorPanics(t *testing.T) {
	const envKey = "APP_INDEXER_SRC_NODES"
	require.NoError(t, os.Setenv(envKey, `not-json`))
	defer os.Unsetenv(envKey)

	v := newTestViper(t, `
indexer:
  chain_id: dorado-1
`)

	require.Panics(t, func() {
		indexerConfig(v)
	})
}

func TestIndexerConfigSrcNodesUnmarshalErrorPanics(t *testing.T) {
	v := viper.New()
	v.Set("indexer.chain_id", "dorado-1")
	v.Set("indexer.src_nodes", []map[string]interface{}{
		{"host": []string{"candidate.example.com"}, "port": "443", "use_tls": true},
	})

	require.Panics(t, func() {
		indexerConfig(v)
	})
}
