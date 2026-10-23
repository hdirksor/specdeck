package update_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/hdirksor/specdeck/internal/update"
)

func releaseServer(t *testing.T, status int, body string) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)
	return server
}

func TestRefresh_StoresLatestRelease(t *testing.T) {
	server := releaseServer(t, http.StatusOK, `{"tag_name":"v0.3.0","html_url":"`+releaseURL+`"}`)
	path := filepath.Join(t.TempDir(), "update-check.json")
	if err := update.SaveState(path, update.State{CheckedAt: now, NotifiedAt: now}); err != nil {
		t.Fatal(err)
	}

	if err := update.Refresh(context.Background(), server.Client(), server.URL, path); err != nil {
		t.Fatalf("Refresh: %v", err)
	}

	got := update.LoadState(path)
	if got.LatestVersion != "v0.3.0" || got.LatestURL != releaseURL {
		t.Errorf("latest = %q %q, want v0.3.0 %s", got.LatestVersion, got.LatestURL, releaseURL)
	}
	if !got.CheckedAt.Equal(now) || !got.NotifiedAt.Equal(now) {
		t.Errorf("Refresh should keep the other fields, got %+v", got)
	}
}

func TestRefresh_FailuresKeepPreviousRelease(t *testing.T) {
	for name, server := range map[string]*httptest.Server{
		"error status": releaseServer(t, http.StatusForbidden, `{"message":"rate limited"}`),
		"bad json":     releaseServer(t, http.StatusOK, `{not json`),
		"missing tag":  releaseServer(t, http.StatusOK, `{"html_url":"`+releaseURL+`"}`),
	} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "update-check.json")
			previous := update.State{LatestVersion: "v0.2.1", LatestURL: "https://example.com/v0.2.1"}
			if err := update.SaveState(path, previous); err != nil {
				t.Fatal(err)
			}

			if err := update.Refresh(context.Background(), server.Client(), server.URL, path); err == nil {
				t.Error("expected an error")
			}
			if got := update.LoadState(path); got != previous {
				t.Errorf("cache changed to %+v", got)
			}
		})
	}
}

func TestRefresh_GivesUpWhenContextExpires(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	t.Cleanup(server.Close)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := update.Refresh(ctx, server.Client(), server.URL, filepath.Join(t.TempDir(), "update-check.json"))
	if err == nil {
		t.Error("expected a timeout error")
	}
}

func TestRefresh_ErrorsOnUnreachableServer(t *testing.T) {
	server := releaseServer(t, http.StatusOK, "")
	url := server.URL
	server.Close()

	if err := update.Refresh(context.Background(), http.DefaultClient, url, filepath.Join(t.TempDir(), "update-check.json")); err == nil {
		t.Error("expected a connection error")
	}
}

func TestRefreshDue(t *testing.T) {
	cases := []struct {
		checkedAt time.Time
		want      bool
	}{
		{time.Time{}, true},
		{now.Add(-25 * time.Hour), true},
		{now.Add(-23 * time.Hour), false},
	}
	for _, c := range cases {
		if got := update.RefreshDue(update.State{CheckedAt: c.checkedAt}, now); got != c.want {
			t.Errorf("CheckedAt %v: RefreshDue = %v, want %v", c.checkedAt, got, c.want)
		}
	}
}
