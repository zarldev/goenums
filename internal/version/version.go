// Package version provides version information for the goenums tool.
// It centralizes version tracking to ensure consistent reporting
// throughout the application.
//
// These values can be accessed directly in code or displayed to users
// via CLI commands. The BUILD and COMMIT constants are designed to be
// populated automatically during the build process using ldflags:
//
//	go build -ldflags="-X github.com/zarldev/goenums/internal/version.BUILD=20230405 -X github.com/zarldev/goenums/internal/version.COMMIT=abc123"
package version

// CURRENT represents the semantic version of the goenums tool.
// This should be manually updated following semantic versioning
var CURRENT string = "v0.3.6"

// BUILD contains build metadata such as the timestamp or build number.
// This field is designed to be populated at build time using the
// -ldflags option:
//
//	-X github.com/zarldev/goenums/internal/version.BUILD=$(date +%Y%m%d-%H:%M:%S)
var BUILD string

// COMMIT contains the git commit hash from which the binary was built.
// This field is designed to be populated at build time using the
// -ldflags option:
//
//	-X github.com/zarldev/goenums/internal/version.COMMIT=$(git rev-parse --short HEAD)
var COMMIT string
