// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cmd

import (
	"testing"

	"github.com/remoterabbit/open-inspector/pkg/model"
)

func TestExitCode(t *testing.T) {
	errd := model.Diagnostics{{Severity: model.SeverityError}}
	warnd := model.Diagnostics{{Severity: model.SeverityWarning}}
	empty := model.Diagnostics{}

	cases := []struct {
		name   string
		diags  model.Diagnostics
		policy FailOnPolicy
		want   int
	}{
		{"empty/error", empty, FailOnError, 0},
		{"warn/error", warnd, FailOnError, 0},
		{"err/error", errd, FailOnError, 2},
		{"warn/warning", warnd, FailOnWarning, 2},
		{"err/warning", errd, FailOnWarning, 2},
		{"err/never", errd, FailOnNever, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ExitCode(tc.diags, tc.policy); got != tc.want {
				t.Errorf("ExitCode = %d, want %d", got, tc.want)
			}
		})
	}
}
