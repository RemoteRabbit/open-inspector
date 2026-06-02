// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Command open-inspector is the CLI front-end for the open-inspector
// library. This scaffold prints a hello-world style banner; subsequent
// steps will introduce cobra subcommands (config, graph, schema).
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/remoterabbit/open-inspector/pkg/inspector"
	"github.com/remoterabbit/open-inspector/pkg/model"
)

func main() {
	jsonOut := flag.Bool("json", false, "emit machine-readable JSON")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: open-inspector [flags] [dir]\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if *showVersion {
		fmt.Println(inspector.Version)
		return
	}

	dir := "."
	if flag.NArg() > 0 {
		dir = flag.Arg(0)
	}

	mod, err := inspector.Inspect(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open-inspector: %v\n", err)
		os.Exit(1)
	}

	if *jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(struct {
			SchemaVersion int           `json:"schema_version"`
			Tool          string        `json:"tool"`
			Version       string        `json:"version"`
			Module        *model.Module `json:"module"`
		}{
			SchemaVersion: model.SchemaVersion,
			Tool:          "open-inspector",
			Version:       inspector.Version,
			Module:        mod,
		}); err != nil {
			fmt.Fprintf(os.Stderr, "open-inspector: encode: %v\n", err)
			os.Exit(1)
		}
		return
	}

	fmt.Printf("open-inspector %s\n", inspector.Version)
	fmt.Printf("inspected: %s\n", mod.Path)
	fmt.Println("(scaffold - config/graph/schema loaders coming soon)")
}
