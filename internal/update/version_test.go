package update_test

import (
	"runtime/debug"
	"testing"

	"github.com/hdirksor/specdeck/internal/update"
)

func buildInfo(version string) *debug.BuildInfo {
	return &debug.BuildInfo{Main: debug.Module{Path: "github.com/hdirksor/specdeck", Version: version}}
}

func TestDetectBuild_StampedVersionIsRelease(t *testing.T) {
	got := update.DetectBuild("v0.2.1", buildInfo("(devel)"))

	want := update.Build{Version: "v0.2.1", Method: update.MethodRelease}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestDetectBuild_CleanModuleVersionIsGoInstall(t *testing.T) {
	got := update.DetectBuild("", buildInfo("v0.2.1"))

	want := update.Build{Version: "v0.2.1", Method: update.MethodGoInstall}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestDetectBuild_LocalBuildsAreDev(t *testing.T) {
	for _, version := range []string{
		"(devel)",
		"",
		"v0.1.1-0.20261023120000-abcdef123456",
		"v0.1.0+dirty",
		"v0.2.0-rc.1",
	} {
		got := update.DetectBuild("", buildInfo(version))
		if got.Method != update.MethodDev {
			t.Errorf("version %q: method = %v, want MethodDev", version, got.Method)
		}
	}
}

func TestDetectBuild_DevKeepsPseudoVersion(t *testing.T) {
	got := update.DetectBuild("", buildInfo("v0.1.1-0.20261023120000-abcdef123456"))

	if got.Version != "v0.1.1-0.20261023120000-abcdef123456" {
		t.Errorf("version = %q, want the pseudo-version", got.Version)
	}
}

func TestDetectBuild_UnknownVersionIsDev(t *testing.T) {
	for _, info := range []*debug.BuildInfo{nil, buildInfo("(devel)"), buildInfo("")} {
		got := update.DetectBuild("", info)
		if got != (update.Build{Version: "dev", Method: update.MethodDev}) {
			t.Errorf("info %+v: got %+v, want dev build", info, got)
		}
	}
}

func TestNewer(t *testing.T) {
	cases := []struct {
		latest, current string
		want            bool
	}{
		{"v0.3.0", "v0.2.1", true},
		{"v0.2.10", "v0.2.9", true},
		{"v1.0.0", "v0.9.9", true},
		{"v0.2.1", "v0.2.1", false},
		{"v0.2.0", "v0.2.1", false},
		{"not-a-version", "v0.2.1", false},
		{"v0.3.0", "dev", false},
	}
	for _, c := range cases {
		if got := update.Newer(c.latest, c.current); got != c.want {
			t.Errorf("Newer(%q, %q) = %v, want %v", c.latest, c.current, got, c.want)
		}
	}
}
