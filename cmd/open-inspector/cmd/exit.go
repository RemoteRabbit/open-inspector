// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cmd

import "github.com/remoterabbit/open-inspector/pkg/model"

// FailOnPolicy translates a Diagnostics slice into an exit code given the user's --fail-on flag value.
type FailOnPolicy string

// Accepted --fail-on policy values.
const (
	FailOnError   FailOnPolicy = "error"
	FailOnWarning FailOnPolicy = "warning"
	FailOnNever   FailOnPolicy = "never"
)

// ExitCode returns the process exit code for the given diagnostics under
// the supplied policy: 2 when a diagnostic matches the threshold, else 0.
//
// TODO: an unrecognized policy (e.g. --fail-on=bogus) falls through to the
// zero return, silently behaving like "never". Validate the --fail-on
// value when the flag is parsed so invalid input errors out instead.
func ExitCode(diags model.Diagnostics, policy FailOnPolicy) int {
	switch policy {
	case FailOnNever:
		return 0
	case FailOnWarning:
		if len(diags) > 0 {
			return 2
		}
	case FailOnError:
		for _, diag := range diags {
			if diag.Severity == model.SeverityError {
				return 2
			}
		}
	}
	return 0
}
