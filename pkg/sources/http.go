// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package sources

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/remoterabbit/open-inspector/pkg/model"
)

func resolveHTTP(parsed ParsedSource, cacheDir string) (model.ResolvedSource, error) {
	hash := cacheKeyHTTP(parsed)
	lock := lockForHash(hash)
	lock.Lock()
	defer lock.Unlock()

	dest := filepath.Join(cacheDir, hash, "extracted")
	if exists(dest) {
		return httpResult(parsed, dest), nil
	}

	tmp := filepath.Join(cacheDir, hash, "tmp")
	_ = os.RemoveAll(tmp)
	if err := os.MkdirAll(tmp, 0o755); err != nil {
		return model.ResolvedSource{}, err
	}

	resp, err := http.Get(parsed.URL)
	if err != nil {
		return model.ResolvedSource{}, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != 200 {
		return model.ResolvedSource{}, fmt.Errorf("http archive: %s returned %d", parsed.URL, resp.StatusCode)
	}

	switch {
	case strings.HasSuffix(parsed.URL, ".zip"):
		err = extractZip(resp.Body, tmp)
	case strings.HasSuffix(parsed.URL, ".tar.gz"), strings.HasSuffix(parsed.URL, ".tgz"):
		err = extractTarGz(resp.Body, tmp)
	case strings.HasSuffix(parsed.URL, ".tar"):
		err = extractTar(resp.Body, tmp)
	default:
		err = fmt.Errorf("unknown archive type: %s", parsed.URL)
	}
	if err != nil {
		return model.ResolvedSource{}, err
	}

	if err := os.Rename(tmp, dest); err != nil {
		return model.ResolvedSource{}, err
	}
	return httpResult(parsed, dest), nil
}

// httpResult builds the ResolvedSource, applying //subdir if present.
func httpResult(parsed ParsedSource, dest string) model.ResolvedSource {
	cachePath := dest
	if parsed.Subdir != "" {
		cachePath = filepath.Join(dest, parsed.Subdir)
	}
	return model.ResolvedSource{
		Kind:      "http",
		Address:   parsed.URL,
		CachePath: cachePath,
	}
}

func extractZip(r io.Reader, dest string) error {
	// ZIP needs a ReaderAt; buffer to memory or temp file.
	buf, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	zipRead, err := zip.NewReader(bytes.NewReader(buf), int64(len(buf)))
	if err != nil {
		return err
	}
	for _, file := range zipRead.File {
		if err := safeExtract(file.Name, dest, func(output io.Writer) error {
			rc, err := file.Open()
			if err != nil {
				return err
			}
			defer func() {
				_ = rc.Close()
			}()
			_, err = io.Copy(output, rc)
			return err
		}); err != nil {
			return err
		}
	}
	return nil
}

// extractTarGz wraps the reader in gzip then walks the tar stream.
func extractTarGz(r io.Reader, dest string) error {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer func() {
		_ = gz.Close()
	}()
	return extractTar(gz, dest)
}

// extractTar streams a tar archive entry-by-entry (no full buffering).
func extractTar(r io.Reader, dest string) error {
	tr := tar.NewReader(r)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		switch header.Typeflag {
		case tar.TypeDir:
			if err := safeExtract(header.Name+"/", dest, nil); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := safeExtract(header.Name, dest, func(out io.Writer) error {
				_, copyErr := io.Copy(out, tr)
				return copyErr
			}); err != nil {
				return err
			}
		}
	}
}

// safeExtract guards against zip/tar slip: any entry whose resolved
// path escapes dest is rejected. A nil write means "directory entry".
func safeExtract(name, dest string, write func(io.Writer) error) error {
	target := filepath.Join(dest, name)
	rel, err := filepath.Rel(dest, target)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("archive entry escapes destination: %q", name)
	}
	if strings.HasSuffix(name, "/") || write == nil {
		return os.MkdirAll(target, 0o755)
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	f, err := os.Create(target)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()
	return write(f)
}
