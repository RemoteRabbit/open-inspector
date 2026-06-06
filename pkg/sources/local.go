// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package sources

import (
	"fmt"
	"path/filepath"

	"github.com/remoterabbit/open-inspector/pkg/model"
)

// resolveLocal resolves a `./`, `../`, or absolute source against the calling module's
// directory and verifies the target exists.
func resolveLocal(parsed ParsedSource, partentDir string) (model.ResolvedSource, error) {
	target := parsed.Raw
	if !filepath.IsAbs(target) {
		target = filepath.Join(partentDir, target)
	}
	target = filepath.Clean(target)
	if !exists(target) {
		return model.ResolvedSource{}, fmt.Errorf("local module path does not exist: %s", target)
	}
	return model.ResolvedSource{
		Kind:      "local",
		Address:   target,
		CachePath: target,
	}, nil
}
