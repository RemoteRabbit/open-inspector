// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package config

import (
	"os"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/remoterabbit/open-inspector/pkg/model"
)

// Listed file kinds and the order Terraform/OpenTofu apply them in.
// Override files (filename ending _override.<ext> or named override.<ext>))
// are collected separately and ignored.

type fileSet struct {
	Primary  []string // .tf, .tf.json, .tofu, .tofu.json (no overrides)
	Override []string // *_override.<ext>, override.<ext>
}

func walk(dir string) (fileSet, model.Diagnostics) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fileSet{}, model.Diagnostics{{
			Severity: model.SeverityError,
			Summary:  "cannot read module directory.",
			Detail:   err.Error(),
			Subject:  &model.Range{Filename: dir},
		}}
	}

	var primary, override []string

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		ext := primaryExt(name)
		if ext == "" {
			continue
		}
		full := dir + string(os.PathSeparator) + name
		if isOverride(name, ext) {
			override = append(override, full)
		} else {
			primary = append(primary, full)
		}
	}
	sort.Strings(primary)
	sort.Strings(override)
	return fileSet{primary, override}, nil
}

// primaryExt returns the longest matching Terraform/OpenTofu file
// extension, or "" if name isn't a config file. Order matters: check
// the two-part extensions first so "foo.tf.json" doesn't get classified
// as ".json".
func primaryExt(name string) string {
	switch {
	case strings.HasSuffix(name, ".tf.json"):
		return ".tf.json"
	case strings.HasSuffix(name, ".tofu.json"):
		return ".tofu.json"
	case strings.HasSuffix(name, ".tf"):
		return ".tf"
	case strings.HasSuffix(name, ".tofu"):
		return ".tofu"
	}
	return ""
}

// isOverride reports whether name is a Terraform override file:
// either exactly "override<ext>" or ending in "_override<ext>".
func isOverride(name, ext string) bool {
	base := strings.TrimSuffix(name, ext)
	return base == "override" || strings.HasSuffix(base, "_override")
}

// silence unused-import warning until parse.go uses hcl directly.
var _ = hcl.DiagError
