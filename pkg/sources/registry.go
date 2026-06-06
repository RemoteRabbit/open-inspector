// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package sources

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"

	"github.com/hashicorp/go-version"
	"github.com/remoterabbit/open-inspector/pkg/model"
)

// registryScheme is "https" in production. Tests override it to "http" to point the
// resolver at an httptest server.
var registryScheme = "https"

// latest returns the highest semver-ordered version, preserving the
// original string form.
func latest(versions []string) string {
	var best *version.Version
	var bestRaw string
	for _, raw := range versions {
		parsed, err := version.NewVersion(raw)
		if err != nil {
			continue
		}
		if best == nil || parsed.GreaterThan(best) {
			best, bestRaw = parsed, raw
		}
	}
	return bestRaw
}

// fetchRegistryVersions returns all published versions for a module.
func fetchRegistryVersions(client *http.Client, host, namespace, name, provider string) ([]string, error) {
	url := fmt.Sprintf("%s://%s/v1/modules/%s/%s/%s/versions",
		registryScheme, host, namespace, name, provider)
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("registry returned %d", resp.StatusCode)
	}
	var payload struct {
		Modules []struct {
			Versions []struct {
				Version string
			} `json:"versions"`
		} `json:"modules"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	// Flatten and return.
	var versions []string
	for _, moduleEntry := range payload.Modules {
		for _, versionEntry := range moduleEntry.Versions {
			versions = append(versions, versionEntry.Version)
		}
	}
	return versions, nil
}

// pickVersion resolves the user's constraint against the published list.
func pickVersion(versions []string, constraint string) (string, error) {
	if constraint == "" {
		return latest(versions), nil
	}
	constraints, err := version.NewConstraint(constraint)
	if err != nil {
		return "", err
	}
	var matches []*version.Version
	for _, candidate := range versions {
		parsed, err := version.NewVersion(candidate)
		if err != nil {
			continue
		}
		if constraints.Check(parsed) {
			matches = append(matches, parsed)
		}
	}
	if len(matches) == 0 {
		return "", fmt.Errorf("no version matches %q", constraint)
	}
	// Pick the highest match.
	var best *version.Version
	for _, match := range matches {
		if best == nil || match.GreaterThan(best) {
			best = match
		}
	}
	return best.Original(), nil
}

// fetchDownloadURL hits the /download endpoint and reads X-Terraform-Get.
// moduleVersion avoids shadowing the imported hashicorp/go-version package.
func fetchDownloadURL(client *http.Client, host, namespace, name, provider, moduleVersion string) (string, error) {
	url := fmt.Sprintf("%s://%s/v1/modules/%s/%s/%s/%s/download",
		registryScheme, host, namespace, name, provider, moduleVersion)
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != 204 {
		return "", fmt.Errorf("registry download returned %d", resp.StatusCode)
	}
	downloadURL := resp.Header.Get("X-Terraform-Get")
	if downloadURL == "" {
		return "", fmt.Errorf("registry response missing X-Terraform-Get header")
	}
	return downloadURL, nil
}

// resolveRegistry resolves namespace/name/provider against the registry
// API, selects a version satisfying constraint, follows X-Terraform-Get
// to the underlying git/http source, and records it under a registry
// cache key.
func resolveRegistry(parsed ParsedSource, constraint, cacheDir string) (model.ResolvedSource, error) {
	client := http.DefaultClient

	versions, err := fetchRegistryVersions(client, parsed.Host, parsed.Namespace, parsed.Name, parsed.Provider)
	if err != nil {
		return model.ResolvedSource{}, err
	}
	selected, err := pickVersion(versions, constraint)
	if err != nil {
		return model.ResolvedSource{}, err
	}

	address := fmt.Sprintf("%s/%s/%s/%s", parsed.Host, parsed.Namespace, parsed.Name, parsed.Provider)
	hash := cacheKeyRegistry(parsed.Host, parsed.Namespace, parsed.Name, parsed.Provider, selected)
	lock := lockForHash(hash)
	lock.Lock()
	defer lock.Unlock()
	if dest := filepath.Join(cacheDir, hash, "extracted"); exists(dest) {
		return model.ResolvedSource{Kind: "registry", Address: address, CachePath: dest, Version: selected}, nil
	}

	downloadURL, err := fetchDownloadURL(client, parsed.Host, parsed.Namespace, parsed.Name, parsed.Provider, selected)
	if err != nil {
		return model.ResolvedSource{}, err
	}

	// X-Terraform-Get is itself a getter-style source; resolve it with
	// the git or http resolver, then re-label it as a registry result.
	inner, err := ParseSource(downloadURL)
	if err != nil {
		return model.ResolvedSource{}, err
	}
	var fetched model.ResolvedSource
	switch inner.Kind {
	case SourceGit:
		fetched, err = resolveGit(inner, cacheDir)
	case SourceHTTP:
		fetched, err = resolveHTTP(inner, cacheDir)
	default:
		return model.ResolvedSource{}, fmt.Errorf("registry returned unsupported download source %q", downloadURL)
	}
	if err != nil {
		return model.ResolvedSource{}, err
	}
	return model.ResolvedSource{
		Kind:      "registry",
		Address:   address,
		CachePath: fetched.CachePath,
		Ref:       fetched.Ref,
		Version:   selected,
	}, nil
}
