// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package sources

import (
	"errors"

	"github.com/remoterabbit/open-inspector/pkg/model"
)

// ErrResolverNotImplemented is returned for source kinds whose resolver has not been wired up yet.
var ErrResolverNotImplemented = errors.New("resolver not implemented for this source kind")

// Resolve parses source, dispatches to the matching resolver, and returns where the module was fetched.
// parentDir is the calling module's directory (for local sources); cacheDir is the cache root for
// fetched (registry/git/http) sources.
//
//nolint:revive // version is consumed by the registry resolver, wired up in a future change.
func Resolve(source, version, parentDir, cacheDir string) (model.ResolvedSource, error) {
	parsed, err := ParseSource(source)
	if err != nil {
		return model.ResolvedSource{}, err
	}
	switch parsed.Kind {
	case SourceLocal:
		return resolveLocal(parsed, parentDir)
	case SourceRegistry:
		return resolveRegistry(parsed, version, cacheDir)
	case SourceGit:
		return resolveGit(parsed, cacheDir)
	case SourceHTTP:
		return resolveHTTP(parsed, cacheDir)
	default:
		return model.ResolvedSource{}, ErrResolverNotImplemented
	}
}
