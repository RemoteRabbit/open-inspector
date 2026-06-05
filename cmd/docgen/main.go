// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Command docgen generates the CLI markdown docs and man pages from the
// cobra command tree.
package main

import (
	"log"
	"time"

	"github.com/remoterabbit/open-inspector/cmd/open-inspector/cmd"
	"github.com/spf13/cobra/doc"
)

// main writes the markdown CLI docs to docs/cli and the man pages to
// docs/man, walking the static cobra command tree exposed by cmd.Root.
func main() {
	root := cmd.Root()

	// Markdown CLI docs.
	if err := doc.GenMarkdownTree(root, "docs/cli"); err != nil {
		log.Fatalf("genmarkdown: %v", err)
	}

	// Man Pages
	header := &doc.GenManHeader{
		Title:   "OPEN-INSPECTOR",
		Section: "1",
		Source:  "open-inspector " + cmd.Version,
		Date:    func() *time.Time { t := time.Now().UTC(); return &t }(),
	}
	if err := doc.GenManTree(root, header, "docs/man"); err != nil {
		log.Fatalf("genman: %v", err)
	}
}
