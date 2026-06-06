// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package sources parses Terraform/OpenTofu module source strings and resolves
// them to a local directory, fetching and caching remote sources as needed.
package sources

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// DefaultCacheDir returns $XDG_CACHE_HOME/open-inspector/modules, falling back to
// %LOCALAPPDATA% on Windows or $HOME/.cache elsewhere. Exported so pkg/inspector
// can seed its default without duplicating the logic.
func DefaultCacheDir() string {
	base := os.Getenv("XDG_CACHE_HOME")
	if base == "" && runtime.GOOS == "windows" {
		base = os.Getenv("LOCALAPPDATA")
	}
	if base == "" {
		if home, err := os.UserHomeDir(); err == nil {
			base = filepath.Join(home, ".cache")
		}
	}
	return filepath.Join(base, "open-inspector", "modules")
}

// cacheKey hashes the canonical identity parts into a hex directory name.
//
//nolint:unused // used by the registry/git/http resolvers, wired up in a future change.
func cacheKey(parts ...string) string {
	sum := sha256.Sum256([]byte(strings.Join(parts, "\x00")))
	return hex.EncodeToString(sum[:])
}

//nolint:unused // used by the git resolver, wired up in a future change.
func cacheKeyGit(parsed ParsedSource) string {
	return cacheKey("git", strings.ToLower(parsed.URL), parsed.Ref)
}

//nolint:unused // used by the http resolver, wired up in a future change.
func cacheKeyHTTP(parsed ParsedSource) string {
	return cacheKey("http", strings.ToLower(parsed.URL))
}

//nolint:unused // used by the registry resolver, wired up in a future change.
func cacheKeyRegistry(host, namespace, name, provider, version string) string {
	return cacheKey("registry", strings.ToLower(host), namespace, name, provider, version)
}

// exists reports whether path exists (file or dir)
func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

var (
	fetchLocksMu sync.Mutex                 //nolint:unused // used by lockForHash, wired up in a future change.
	fetchLocks   = map[string]*sync.Mutex{} //nolint:unused // used by lockForHash, wired up in a future change.
)

// lockForHash returns a per-hash mutex so concurrent resolves of the same module
// serialize their fetch/extract instead of racing.
//
//nolint:unused // used by the registry/git/http resolvers, wired up in a future change.
func lockForHash(hash string) *sync.Mutex {
	fetchLocksMu.Lock()
	defer fetchLocksMu.Unlock()
	mutex, ok := fetchLocks[hash]
	if !ok {
		mutex = &sync.Mutex{}
		fetchLocks[hash] = mutex
	}
	return mutex
}
