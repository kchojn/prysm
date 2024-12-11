package params_test

import (
	"path"
	"testing"

	"github.com/bazelbuild/rules_go/go/tools/bazel"
	"github.com/prysmaticlabs/prysm/v5/config/params"
	"github.com/prysmaticlabs/prysm/v5/testing/require"
)

func TestSepoliaConfigMatchesUpstreamYaml(t *testing.T) {
	presetFPs := presetsFilePath(t, "mainnet")
	mn, err := params.ByName(params.MainnetName)
	require.NoError(t, err)
	cfg := mn.Copy()
	for _, fp := range presetFPs {
		cfg, err = params.UnmarshalConfigFile(fp, cfg)
		require.NoError(t, err)
	}
	fPath, err := bazel.Runfile("external/sepolia_testnet")
	require.NoError(t, err)
	configFP := path.Join(fPath, "metadata", "config.yaml")
	pcfg, err := params.UnmarshalConfigFile(configFP, nil)
	require.NoError(t, err)
	fields := fieldsFromYamls(t, append(presetFPs, configFP))
	// TODO(#14714): remove hard coded sizes once spec is updated
	pcfg.MaxChunkSize = 15728640
	pcfg.GossipMaxSize = 15728640
	assertYamlFieldsMatch(t, "sepolia", fields, pcfg, params.SepoliaConfig())
}
