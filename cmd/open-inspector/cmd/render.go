// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/remoterabbit/open-inspector/pkg/inspector"
	"github.com/remoterabbit/open-inspector/pkg/model"
)

// jsonEnvelope is the stable, versioned JSON output shape.
type jsonEnvelope struct {
	SchemaVersion int           `json:"schema_version"`
	Tool          string        `json:"tool"`
	Version       string        `json:"version"`
	Module        *model.Module `json:"module"`
}

func renderJSON(cmd *cobra.Command, mod *model.Module) error {
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	if err := enc.Encode(jsonEnvelope{
		SchemaVersion: model.SchemaVersion,
		Tool:          "open-inspector",
		Version:       inspector.Version,
		Module:        mod,
	}); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	return nil
}

// errWriter wraps an io.Writer and remembers the first write error so
// callers can check once at the end instead of after every write.
type errWriter struct {
	w   io.Writer
	err error
}

func (ew *errWriter) printf(format string, args ...any) {
	if ew.err != nil {
		return
	}
	_, ew.err = fmt.Fprintf(ew.w, format, args...)
}

// renderTable prints a grouped, human-readable summary of the module.
// Empty sections are omitted so we never print a zero-row header.
func renderTable(cmd *cobra.Command, mod *model.Module) error {
	ew := &errWriter{w: cmd.OutOrStdout()}
	ew.printf("# %s\n\n", mod.Path)
	renderVariables(ew, mod.Variables)
	renderOutputs(ew, mod.Outputs)
	renderLocals(ew, mod.Locals)
	renderResources(ew, "Managed resources", mod.ManagedResources)
	renderResources(ew, "Data resources", mod.DataResources)
	renderModuleCalls(ew, mod.ModuleCalls)
	renderProviders(ew, mod.Providers)
	renderRequiredProviders(ew, mod.RequiredProviders)
	renderDiagnostics(ew, mod.Diagnostics)
	if ew.err != nil {
		return fmt.Errorf("write table: %w", ew.err)
	}
	return nil
}

func renderVariables(ew *errWriter, vars []model.Variable) {
	if len(vars) == 0 {
		return
	}
	ew.printf("## Variables (%d)\n", len(vars))
	rows := [][]string{{"NAME", "TYPE", "DEFAULT", "SENSITIVE", "DESCRIPTION"}}
	for _, v := range vars {
		def := "(none)"
		if v.Default != nil {
			def = truncate(v.Default.Source, 30)
		}
		rows = append(rows, []string{
			v.Name, truncate(orNone(v.Type), 24), def, yesNo(v.Sensitive), orNone(v.Description),
		})
	}
	ew.table(rows)
}

func renderOutputs(ew *errWriter, outs []model.Output) {
	if len(outs) == 0 {
		return
	}
	ew.printf("## Outputs (%d)\n", len(outs))
	rows := [][]string{{"NAME", "SENSITIVE", "DESCRIPTION"}}
	for _, o := range outs {
		rows = append(rows, []string{o.Name, yesNo(o.Sensitive), orNone(o.Description)})
	}
	ew.table(rows)
}

func renderLocals(ew *errWriter, locals []model.Local) {
	if len(locals) == 0 {
		return
	}
	ew.printf("## Locals (%d)\n", len(locals))
	rows := [][]string{{"NAME", "VALUE (source)"}}
	for _, l := range locals {
		rows = append(rows, []string{l.Name, truncate(l.Value.Source, 60)})
	}
	ew.table(rows)
}

func renderResources(ew *errWriter, heading string, res []model.Resource) {
	if len(res) == 0 {
		return
	}
	ew.printf("## %s (%d)\n", heading, len(res))
	rows := [][]string{{"TYPE", "NAME", "PROVIDER", "COUNT/FOR_EACH"}}
	for _, r := range res {
		rows = append(rows, []string{
			r.Type, r.Name, orNone(r.Provider), repetition(r.Count, r.ForEach),
		})
	}
	ew.table(rows)
}

func renderModuleCalls(ew *errWriter, calls []model.ModuleCall) {
	if len(calls) == 0 {
		return
	}
	ew.printf("## Module calls (%d)\n", len(calls))
	rows := [][]string{{"NAME", "SOURCE", "VERSION"}}
	for _, c := range calls {
		rows = append(rows, []string{c.Name, c.Source, orNone(c.Version)})
	}
	ew.table(rows)
}

func renderProviders(ew *errWriter, providers []model.ProviderConfig) {
	if len(providers) == 0 {
		return
	}
	ew.printf("## Providers (%d)\n", len(providers))
	rows := [][]string{{"NAME", "ALIAS"}}
	for _, p := range providers {
		rows = append(rows, []string{p.Name, orNone(p.Alias)})
	}
	ew.table(rows)
}

func renderRequiredProviders(ew *errWriter, reqs map[string]model.ProviderRequirement) {
	if len(reqs) == 0 {
		return
	}
	names := make([]string, 0, len(reqs))
	for name := range reqs {
		names = append(names, name)
	}
	sort.Strings(names)
	ew.printf("## Required providers (%d)\n", len(reqs))
	rows := [][]string{{"NAME", "SOURCE", "VERSION"}}
	for _, name := range names {
		req := reqs[name]
		rows = append(rows, []string{
			name, orNone(req.Source), orNone(strings.Join(req.VersionConstraints, ", ")),
		})
	}
	ew.table(rows)
}

func renderDiagnostics(ew *errWriter, diags model.Diagnostics) {
	if len(diags) == 0 {
		return
	}
	ew.printf("## Diagnostics (%d)\n", len(diags))
	rows := [][]string{{"SEVERITY", "SUMMARY"}}
	for _, d := range diags {
		rows = append(rows, []string{string(d.Severity), truncate(d.Summary, 70)})
	}
	ew.table(rows)
}

// table renders rows (the first row is the header) as an aligned,
// tab-separated block followed by a trailing blank line.
func (ew *errWriter) table(rows [][]string) {
	if ew.err != nil {
		return
	}
	tw := tabwriter.NewWriter(ew.w, 0, 0, 2, ' ', 0)
	for _, row := range rows {
		if _, err := fmt.Fprintln(tw, strings.Join(row, "\t")); err != nil {
			ew.err = err
			return
		}
	}
	if err := tw.Flush(); err != nil {
		ew.err = err
		return
	}
	ew.printf("\n")
}

func truncate(s string, n int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) <= n {
		return s
	}
	if n <= 3 {
		return s[:n]
	}
	return s[:n-3] + "..."
}

func orNone(s string) string {
	if s == "" {
		return "(none)"
	}
	return s
}

func yesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

func repetition(count, forEach *model.Expression) string {
	switch {
	case count != nil:
		return "count=" + truncate(count.Source, 20)
	case forEach != nil:
		return "for_each=" + truncate(forEach.Source, 20)
	default:
		return "(none)"
	}
}
