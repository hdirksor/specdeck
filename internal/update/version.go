// Package update tells users when a newer specdeck release is available.
package update

import (
	"regexp"
	"runtime/debug"
	"strconv"
)

// Method is how the running binary was installed.
type Method int

const (
	// MethodDev is a local build; update checks are skipped.
	MethodDev Method = iota
	// MethodGoInstall is a `go install` of a tagged release.
	MethodGoInstall
	// MethodRelease is a binary downloaded from a GitHub release.
	MethodRelease
)

// Build describes the running binary.
type Build struct {
	Version string
	Method  Method
}

var releaseVersion = regexp.MustCompile(`^v(\d+)\.(\d+)\.(\d+)$`)

// DetectBuild works out the version and install method. stamped is the
// version goreleaser sets at link time; info is Go's embedded build info and
// may be nil.
func DetectBuild(stamped string, info *debug.BuildInfo) Build {
	if stamped != "" {
		return Build{Version: stamped, Method: MethodRelease}
	}
	if info == nil || info.Main.Version == "" || info.Main.Version == "(devel)" {
		return Build{Version: "dev", Method: MethodDev}
	}
	if releaseVersion.MatchString(info.Main.Version) {
		return Build{Version: info.Main.Version, Method: MethodGoInstall}
	}
	return Build{Version: info.Main.Version, Method: MethodDev}
}

// Newer reports whether latest is a later release than current. Anything that
// is not a plain vX.Y.Z release is never newer.
func Newer(latest, current string) bool {
	latestParts, ok := parseVersion(latest)
	if !ok {
		return false
	}
	currentParts, ok := parseVersion(current)
	if !ok {
		return false
	}
	for i := range latestParts {
		if latestParts[i] != currentParts[i] {
			return latestParts[i] > currentParts[i]
		}
	}
	return false
}

func parseVersion(version string) ([3]int, bool) {
	match := releaseVersion.FindStringSubmatch(version)
	if match == nil {
		return [3]int{}, false
	}
	var parts [3]int
	for i := range parts {
		n, err := strconv.Atoi(match[i+1])
		if err != nil {
			return [3]int{}, false
		}
		parts[i] = n
	}
	return parts, true
}
