package version

import (
	"fmt"
)

const Major = "0"
const Minor = "1"
const Patch = "0"
const Label = "rc.1"

var (
	// Version is the current version of Travis
	Version string

	// GitCommit is set with --ldflags "-X main.gitCommit=$(git rev-parse --short HEAD)"
	GitCommit string
)

func init() {
	Version = fmt.Sprintf("%s.%s.%s", Major, Minor, Patch)

	if Label != "" {
		Version += "-" + Label
	}

	if GitCommit != "" {
		Version += "-" + GitCommit
	}
}
