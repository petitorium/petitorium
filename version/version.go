// Package version holds the build-time-injected version of Petitorium.
package version

// Version is the current version of Petitorium.
// It is set at build time using -ldflags by GoReleaser; defaults to "dev"
// for local builds that do not pass ldflags.
var Version = "dev"
