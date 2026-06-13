// Command rollops-plugin-split is a Rollops feature-flag provider plugin backed
// by Split (Harness FME). Build it, pin its sha256, and point a rollout's
// featureFlags.plugin at the binary.
package main

import (
	"fmt"
	"os"

	split "github.com/klarlabs-studio/rollops-plugin-split"
	"go.klarlabs.de/rollops/pkg/plugin"
)

// version is overwritten at build time via -ldflags.
var version = "dev"

func main() {
	safety := plugin.Safety{
		NetworkHosts: []string{"api.split.io:443"},
		EnvVars:      []string{"SPLIT_API_URL", "SPLIT_TOKEN"},
		RiskClass:    plugin.RiskActive,
	}
	if err := plugin.ServeFlagProvider("klarlabs/split", version, split.FromEnv(), safety); err != nil {
		fmt.Fprintln(os.Stderr, "rollops-plugin-split:", err)
		os.Exit(1)
	}
}
