// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package config

import (
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/remoterabbit/open-inspector/pkg/model"
)

type parsedFiles struct {
	files  []*hcl.File
	parser *hclparse.Parser
}

func parse(fs fileSet) (parsedFiles, model.Diagnostics) {
	parse := hclparse.NewParser()
	var diags hcl.Diagnostics
	var files []*hcl.File

	for _, path := range fs.Primary {
		var file *hcl.File
		var diag hcl.Diagnostics

		switch {
		case strings.HasSuffix(path, ".tf.json"),
			strings.HasSuffix(path, ".tofu.json"):
			file, diag = parse.ParseJSONFile(path)
		default:
			file, diag = parse.ParseHCLFile(path)
		}
		diags = append(diags, diag...)
		if file != nil {
			files = append(files, file)
		}
	}
	return parsedFiles{files: files, parser: parse}, model.DiagnosticsFromHCL(diags)
}
