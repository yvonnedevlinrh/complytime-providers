// SPDX-License-Identifier: Apache-2.0

// Package version provides build-time version injection for all
// provider binaries via -ldflags.
package version

import "strings"

// version is set at build time via -ldflags
// "-X github.com/complytime/complytime-providers/internal/version.version=<value>".
var version string

// Version returns the build version string. It strips a leading "v" prefix
// if present to normalize tag-style versions (e.g., "v0.1.0" -> "0.1.0").
// Returns "0.0.0-unknown" when no version was injected at build time.
func Version() string {
	if version == "" {
		return "0.0.0-unknown"
	}
	return strings.TrimPrefix(version, "v")
}
