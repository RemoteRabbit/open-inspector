// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package sources

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestHTTPResolver_Zip(t *testing.T) {
	buf := buildZip(t, map[string]string{
		"main.tf": `resource "null_resource" "x" {}`,
	})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(buf)
	}))
	defer srv.Close()

	resolved, err := resolveHTTP(ParsedSource{
		Kind: SourceHTTP, URL: srv.URL + "/m.zip",
	}, t.TempDir())
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if !exists(filepath.Join(resolved.CachePath, "main.tf")) {
		t.Errorf("main.tf missing from %s", resolved.CachePath)
	}
}

func TestHTTPResolver_TarGz(t *testing.T) {
	buf := buildTarGz(t, map[string]string{
		"sub/main.tf": `resource "null_resource" "x" {}`,
	})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(buf)
	}))
	defer srv.Close()

	for _, suffix := range []string{"/m.tar.gz", "/m.tgz"} {
		resolved, err := resolveHTTP(ParsedSource{
			Kind: SourceHTTP, URL: srv.URL + suffix,
		}, t.TempDir())
		if err != nil {
			t.Fatalf("resolve %s: %v", suffix, err)
		}
		if !exists(filepath.Join(resolved.CachePath, "sub", "main.tf")) {
			t.Errorf("%s: sub/main.tf missing from %s", suffix, resolved.CachePath)
		}
	}
}

func TestExtractTar_RejectsEscape(t *testing.T) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	body := []byte("pwned")
	if err := tw.WriteHeader(&tar.Header{
		Name: "../escape.tf", Mode: 0o644, Size: int64(len(body)), Typeflag: tar.TypeReg,
	}); err != nil {
		t.Fatalf("tar header: %v", err)
	}
	if _, err := tw.Write(body); err != nil {
		t.Fatalf("tar write: %v", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("tar close: %v", err)
	}

	dest := t.TempDir()
	if err := extractTar(bytes.NewReader(buf.Bytes()), dest); err == nil {
		t.Fatalf("expected tar-slip rejection, got nil error")
	}
	if exists(filepath.Join(filepath.Dir(dest), "escape.tf")) {
		t.Errorf("tar-slip entry escaped the destination directory")
	}
}

func TestExtractZip_RejectsEscape(t *testing.T) {
	buf := buildZip(t, map[string]string{
		"../escape.tf": "pwned",
	})
	dest := t.TempDir()
	if err := extractZip(bytes.NewReader(buf), dest); err == nil {
		t.Fatalf("expected zip-slip rejection, got nil error")
	}
	if exists(filepath.Join(filepath.Dir(dest), "escape.tf")) {
		t.Errorf("zip-slip entry escaped the destination directory")
	}
}

// buildTarGz returns the bytes of a gzip-compressed tar archive
// containing the given path -> contents entries.
func buildTarGz(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for name, body := range files {
		if err := tw.WriteHeader(&tar.Header{
			Name: name, Mode: 0o644, Size: int64(len(body)), Typeflag: tar.TypeReg,
		}); err != nil {
			t.Fatalf("tar header %s: %v", name, err)
		}
		if _, err := tw.Write([]byte(body)); err != nil {
			t.Fatalf("tar write %s: %v", name, err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("tar close: %v", err)
	}
	if err := gz.Close(); err != nil {
		t.Fatalf("gzip close: %v", err)
	}
	return buf.Bytes()
}

// buildZip returns the bytes of a zip archive containing the given
// path -> contents entries.
func buildZip(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, body := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatalf("zip create %s: %v", name, err)
		}
		if _, err := w.Write([]byte(body)); err != nil {
			t.Fatalf("zip write %s: %v", name, err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("zip close: %v", err)
	}
	return buf.Bytes()
}
