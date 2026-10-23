package build_test

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/hdirksor/specdeck/internal/build"
)

func writeDist(t *testing.T, containers ...build.Container) (string, []string) {
	t.Helper()
	distRoot := filepath.Join(t.TempDir(), "dist")
	written, err := build.WriteDist(containers, t.TempDir(), distRoot)
	if err != nil {
		t.Fatalf("WriteDist: %v", err)
	}
	return distRoot, written
}

func assertNotExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("expected %s not to exist", path)
	}
}

func TestWriteDist_TopLevelLeaf(t *testing.T) {
	distRoot, written := writeDist(t, build.Container{Path: "home.yml", Title: "Home"})

	assertContains(t, readFile(t, filepath.Join(distRoot, "home.md")), "# Home")
	if !slices.Equal(written, []string{"home.md"}) {
		t.Errorf("written = %v, want [home.md]", written)
	}
}

func TestWriteDist_TopLevelIndex(t *testing.T) {
	distRoot, written := writeDist(t, build.Container{Path: "index.yml", Title: "Root"})

	assertContains(t, readFile(t, filepath.Join(distRoot, "_index.md")), "# Root")
	assertNotExists(t, filepath.Join(distRoot, "index.md"))
	if !slices.Equal(written, []string{"_index.md"}) {
		t.Errorf("written = %v, want [_index.md]", written)
	}
}

func TestWriteDist_NestedIndex(t *testing.T) {
	distRoot, _ := writeDist(t, build.Container{Path: "settings/index.yml", Title: "Settings"})

	assertContains(t, readFile(t, filepath.Join(distRoot, "settings", "_index.md")), "# Settings")
}

func TestWriteDist_NestedLeafWithoutIndexGetsSectionStub(t *testing.T) {
	distRoot, written := writeDist(t, build.Container{Path: "user-profile/card.yml", Title: "Card"})

	assertContains(t, readFile(t, filepath.Join(distRoot, "user-profile", "card.md")), "# Card")
	stub := readFile(t, filepath.Join(distRoot, "user-profile", "_index.md"))
	if stub != "---\ntitle: User Profile\n---\n" {
		t.Errorf("stub = %q", stub)
	}
	want := []string{"user-profile/card.md", "user-profile/_index.md"}
	if !slices.Equal(written, want) {
		t.Errorf("written = %v, want %v", written, want)
	}
}

func TestWriteDist_NestedLeafWithIndexKeepsIndex(t *testing.T) {
	distRoot, written := writeDist(t,
		build.Container{Path: "settings/index.yml", Title: "Settings"},
		build.Container{Path: "settings/privacy.yml", Title: "Privacy"},
	)

	index := readFile(t, filepath.Join(distRoot, "settings", "_index.md"))
	assertContains(t, index, "# Settings")
	assertNotContains(t, index, "title: Settings")
	want := []string{"settings/_index.md", "settings/privacy.md"}
	if !slices.Equal(written, want) {
		t.Errorf("written = %v, want %v", written, want)
	}
}

func TestWriteDist_IndexAfterLeafStillSuppressesStub(t *testing.T) {
	distRoot, _ := writeDist(t,
		build.Container{Path: "settings/privacy.yml", Title: "Privacy"},
		build.Container{Path: "settings/index.yml", Title: "Settings"},
	)

	assertContains(t, readFile(t, filepath.Join(distRoot, "settings", "_index.md")), "# Settings")
}

func TestWriteDist_DeeplyNestedStubPathIsFullPath(t *testing.T) {
	distRoot, written := writeDist(t, build.Container{Path: "account/billing/card.yml", Title: "Card"})

	stub := readFile(t, filepath.Join(distRoot, "account", "billing", "_index.md"))
	assertContains(t, stub, "title: Billing")
	if !slices.Contains(written, "account/billing/_index.md") {
		t.Errorf("written = %v, want it to contain account/billing/_index.md", written)
	}
}

func TestWriteDist_StubsAreSorted(t *testing.T) {
	_, written := writeDist(t,
		build.Container{Path: "zebra/a.yml", Title: "A"},
		build.Container{Path: "alpha/b.yml", Title: "B"},
		build.Container{Path: "middle/c.yml", Title: "C"},
	)

	want := []string{"zebra/a.md", "alpha/b.md", "middle/c.md",
		"alpha/_index.md", "middle/_index.md", "zebra/_index.md"}
	if !slices.Equal(written, want) {
		t.Errorf("written = %v, want %v", written, want)
	}
}

func TestWriteDist_StubTitleSplitsHyphensAndUnderscores(t *testing.T) {
	distRoot, _ := writeDist(t, build.Container{Path: "user-profile_page/card.yml", Title: "Card"})

	assertContains(t, readFile(t, filepath.Join(distRoot, "user-profile_page", "_index.md")), "title: User Profile Page")
}

func TestWriteDist_CopiesImageNextToOutput(t *testing.T) {
	containersRoot := t.TempDir()
	pngData := encodePNG(t, 10, 10)
	writeFile(t, filepath.Join(containersRoot, "screens", "index.png"), string(pngData))
	distRoot := filepath.Join(t.TempDir(), "dist")

	c := build.Container{Path: filepath.Join("screens", "index.yml"), Title: "Screens", Image: "index.png"}
	if _, err := build.WriteDist([]build.Container{c}, containersRoot, distRoot); err != nil {
		t.Fatalf("WriteDist: %v", err)
	}

	if got := readFile(t, filepath.Join(distRoot, "screens", "index.png")); got != string(pngData) {
		t.Errorf("copied image: want %d bytes, got %d bytes", len(pngData), len(got))
	}
}
