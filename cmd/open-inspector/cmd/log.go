// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cmd

import (
	"fmt"
	"io"
	"os"
	"regexp"
)

// ansiRE matches ANSI SGR (color) escape sequences so they can be
// stripped when color output is disabled.
var ansiRE = regexp.MustCompile("\x1b\\[[0-9;]*m")

// logger writes informational messages to a destination (stderr by
// default), honoring the --quiet and --no-color flags.
type logger struct {
	w       io.Writer
	quiet   bool
	noColor bool
}

// newLogger builds a logger. Color is disabled when noColor is set or the
// NO_COLOR environment variable is present (any value).
func newLogger(w io.Writer, quiet, noColor bool) *logger {
	_, noColorEnv := os.LookupEnv("NO_COLOR")
	return &logger{w: w, quiet: quiet, noColor: noColor || noColorEnv}
}

// infof writes an informational message followed by a newline. It is a
// no-op when quiet is set; ANSI color codes are stripped when color is
// disabled.
func (l *logger) infof(format string, args ...any) {
	if l.quiet {
		return
	}
	msg := fmt.Sprintf(format, args...)
	if l.noColor {
		msg = ansiRE.ReplaceAllString(msg, "")
	}
	// Best effort; an unwritable log destination is not fatal.
	_, _ = fmt.Fprintln(l.w, msg)
}

// stderrLog is the process-wide logger, configured from the global flags
// once cobra has parsed them (see root.go's cobra.OnInitialize).
var stderrLog = newLogger(os.Stderr, false, false)
