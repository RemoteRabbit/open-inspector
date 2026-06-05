// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Command open-inspector is the CLI entrypoint. It delegates to the cobra
// command tree in the cmd package and exits with the returned status code.
package main

import (
	"os"

	"github.com/remoterabbit/open-inspector/cmd/open-inspector/cmd"
)

// main runs the root command and propagates its exit code to the process.
func main() {
	os.Exit(cmd.Execute())
}
