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

// parsedFiles holds the parsed HCL files for a module, keeping primary and
// override files separate, plus the parser that owns their source bytes.
type parsedFiles struct {
	primary  []*hcl.File
	override []*hcl.File
	parser   *hclparse.Parser
}

// parse parses every file in fileSet, dispatching on extension between the
// JSON and native-HCL parsers. The returned parser retains the source
// bytes needed later for verbatim expression capture.
func parse(fileSet fileSet) (parsedFiles, model.Diagnostics) {
	parser := hclparse.NewParser()
	var diags hcl.Diagnostics

	parseAll := func(paths []string) []*hcl.File {
		var files []*hcl.File

		for _, path := range paths {
			var file *hcl.File
			var diag hcl.Diagnostics

			switch {
			case strings.HasSuffix(path, ".tf.json"),
				strings.HasSuffix(path, ".tofu.json"):
				file, diag = parser.ParseJSONFile(path)
			default:
				file, diag = parser.ParseHCLFile(path)
			}
			diags = append(diags, diag...)

			if file != nil {
				files = append(files, file)
			}
		}
		return files
	}

	return parsedFiles{
		primary:  parseAll(fileSet.Primary),
		override: parseAll(fileSet.Override),
		parser:   parser,
	}, model.DiagnosticsFromHCL(diags)
}
