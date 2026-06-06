// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package sources

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRegistryResolver(t *testing.T) {
	zipBytes := buildZip(t, map[string]string{
		"main.tf": `resource "null_resource" "x" {}`,
	})
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/modules/test/foo/aws/versions", func(w http.ResponseWriter, _ *http.Request) {
		if err := json.NewEncoder(w).Encode(map[string]any{
			"modules": []map[string]any{{
				"versions": []map[string]string{
					{"version": "1.0.0"}, {"version": "1.1.0"}, {"version": "2.0.0-alpha"},
				},
			}},
		}); err != nil {
			t.Errorf("encode versions response: %v", err)
		}
	})
	mux.HandleFunc("/m.zip", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(zipBytes)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	mux.HandleFunc("/v1/modules/test/foo/aws/1.1.0/download", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Terraform-Get", srv.URL+"/m.zip")
		w.WriteHeader(204)
	})

	// Pure version selection needs no server.
	got, err := pickVersion([]string{"1.0.0", "1.1.0", "2.0.0-alpha"}, "~> 1.0")
	if err != nil {
		t.Fatalf("pickVersion: %v", err)
	}
	if got != "1.1.0" {
		t.Errorf("pickVersion(~>1.0): got %q, want 1.1.0", got)
	}

	// End-to-end against the test server: override scheme + host so the
	// resolver hits httptest instead of the real registry.
	defer func(prev string) { registryScheme = prev }(registryScheme)
	registryScheme = "http"
	host := strings.TrimPrefix(srv.URL, "http://") // "127.0.0.1:PORT"

	parsed := ParsedSource{
		Kind: SourceRegistry, Host: host,
		Namespace: "test", Name: "foo", Provider: "aws",
	}
	// Full round-trip: version listing, constraint solving, download URL
	// fetch, dispatch to the http resolver, and zip extraction.
	resolved, err := resolveRegistry(parsed, "~> 1.0", t.TempDir())
	if err != nil {
		t.Fatalf("resolveRegistry: %v", err)
	}
	if resolved.Kind != "registry" {
		t.Errorf("resolved.Kind: got %q, want registry", resolved.Kind)
	}
	if resolved.Version != "1.1.0" {
		t.Errorf("resolved.Version: got %q, want 1.1.0", resolved.Version)
	}
	if _, err := os.Stat(filepath.Join(resolved.CachePath, "main.tf")); err != nil {
		t.Errorf("expected main.tf in cache path: %v", err)
	}
}

func TestPickVersion_EmptyConstraintPicksLatest(t *testing.T) {
	got, err := pickVersion([]string{"1.0.0", "2.3.1", "1.5.0"}, "")
	if err != nil {
		t.Fatalf("pickVersion: %v", err)
	}
	if got != "2.3.1" {
		t.Errorf("pickVersion(empty constraint): got %q, want 2.3.1", got)
	}
}

func TestLatest(t *testing.T) {
	if got := latest([]string{"0.9.0", "1.10.0", "1.2.0", "bogus"}); got != "1.10.0" {
		t.Errorf("latest: got %q, want 1.10.0", got)
	}
	if got := latest(nil); got != "" {
		t.Errorf("latest(nil): got %q, want empty", got)
	}
}
