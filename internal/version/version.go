package version

import "fmt"

// Version is the version number of Kubectl-Confirm
var Version = ""

// GitCommit is the git commit hash of Kubectl-Confirm
var GitCommit = ""

// String returns a formatted version string
func String() string {
	version := Version
	if version == "" {
		version = "UNKNOWN"
	}
	gitCommit := GitCommit
	if gitCommit == "" {
		gitCommit = "UNKNOWN"
	}
	return fmt.Sprintf("%s (git commit %s)", version, gitCommit)
}
