package main

// Build metadata - these variables are set at build time via -ldflags
var (
	// GitCommit is the git commit hash
	GitCommit = "dev"
	// BuildDate is the build date in RFC3339 format
	BuildDate = "unknown"
)
