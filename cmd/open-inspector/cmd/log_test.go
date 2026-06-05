// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestLogger_Quiet(t *testing.T) {
	var buf bytes.Buffer
	l := newLogger(&buf, true, false)
	l.infof("hello %s", "world")
	if buf.Len() != 0 {
		t.Errorf("quiet logger wrote output: %q", buf.String())
	}
}

func TestLogger_Info(t *testing.T) {
	var buf bytes.Buffer
	l := newLogger(&buf, false, false)
	l.infof("hello %s", "world")
	if got := buf.String(); got != "hello world\n" {
		t.Errorf("infof = %q, want %q", got, "hello world\n")
	}
}

func TestLogger_StripsANSIWhenNoColor(t *testing.T) {
	var buf bytes.Buffer
	l := newLogger(&buf, false, true)
	l.infof("\x1b[31mred\x1b[0m text")
	got := buf.String()
	if strings.Contains(got, "\x1b") {
		t.Errorf("no-color logger left ANSI codes: %q", got)
	}
	if got != "red text\n" {
		t.Errorf("infof = %q, want %q", got, "red text\n")
	}
}

func TestLogger_NoColorEnvStripsANSI(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	var buf bytes.Buffer
	l := newLogger(&buf, false, false)
	l.infof("\x1b[31mred\x1b[0m")
	if got := buf.String(); strings.Contains(got, "\x1b") {
		t.Errorf("NO_COLOR env did not disable color: %q", got)
	}
}
