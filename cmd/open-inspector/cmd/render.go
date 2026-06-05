// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
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

// renderJSON writes the module as an indented, versioned JSON envelope to
// the command's output stream.
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

// printf writes a formatted string, recording the first error and
// skipping all subsequent writes once one has occurred.
func (ew *errWriter) printf(format string, args ...any) {
	if ew.err != nil {
		return
	}
	_, ew.err = fmt.Fprintf(ew.w, format, args...)
}

// renderTable prints a grouped, human-readable summary of the module.
// Empty sections are omitted so we never print a zero-row header. The
// table favors a scannable overview: nested detail (validation
// conditions, lifecycle bodies, encryption bodies, source ranges) is
// summarized as presence/counts here and is available in full via --json.
func renderTable(cmd *cobra.Command, mod *model.Module) error {
	ew := &errWriter{w: cmd.OutOrStdout()}
	ew.printf("# %s\n\n", mod.Path)
	renderRequiredCore(ew, mod.RequiredCore)
	renderRequiredProviders(ew, mod.RequiredProviders)
	renderVariables(ew, mod.Variables)
	renderOutputs(ew, mod.Outputs)
	renderLocals(ew, mod.Locals)
	renderResources(ew, "Managed resources", mod.ManagedResources)
	renderResources(ew, "Data resources", mod.DataResources)
	renderEphemeralResources(ew, mod.EphemeralResources)
	renderModuleCalls(ew, mod.ModuleCalls)
	renderProviders(ew, mod.Providers)
	renderMoved(ew, mod.Moved)
	renderImports(ew, mod.Imports)
	renderRemoved(ew, mod.Removed)
	renderChecks(ew, mod.Checks)
	renderEncryption(ew, mod.Encryption)
	renderDiagnostics(ew, mod.Diagnostics)
	ew.printf("(full detail available with --json)\n")
	if ew.err != nil {
		return fmt.Errorf("write table: %w", ew.err)
	}
	return nil
}

// renderRequiredCore prints the terraform/tofu required_version
// constraints. Omitted when none are declared.
func renderRequiredCore(ew *errWriter, constraints []string) {
	if len(constraints) == 0 {
		return
	}
	ew.printf("## Required core (%d)\n", len(constraints))
	ew.printf("%s\n\n", strings.Join(constraints, ", "))
}

// renderVariables prints the variable declarations. Validation blocks are
// summarized as a count; their conditions live in the JSON output.
func renderVariables(ew *errWriter, vars []model.Variable) {
	if len(vars) == 0 {
		return
	}
	ew.printf("## Variables (%d)\n", len(vars))
	rows := [][]string{{"NAME", "TYPE", "DEFAULT", "SENSITIVE", "VALIDATIONS", "DESCRIPTION"}}
	for _, v := range vars {
		def := "(none)"
		if v.Default != nil {
			def = truncate(v.Default.Source, 30)
		}
		rows = append(rows, []string{
			v.Name, truncate(orNone(v.Type), 24), def, yesNo(v.Sensitive),
			count(len(v.Validations)), orNone(v.Description),
		})
	}
	ew.table(rows)
}

// renderOutputs prints the output declarations. Omitted when none exist.
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

// renderLocals prints the local values with their source expressions.
// Omitted when none exist.
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

// renderResources prints a resource section (managed or data) under the
// given heading. Lifecycle blocks are reduced to a compact summary; the
// full detail is in the JSON output. Omitted when the slice is empty.
func renderResources(ew *errWriter, heading string, res []model.Resource) {
	if len(res) == 0 {
		return
	}
	ew.printf("## %s (%d)\n", heading, len(res))
	rows := [][]string{{"TYPE", "NAME", "PROVIDER", "COUNT/FOR_EACH", "LIFECYCLE"}}
	for _, r := range res {
		rows = append(rows, []string{
			r.Type, r.Name, orNone(r.Provider), repetition(r.Count, r.ForEach),
			lifecycleSummary(r.Lifecycle),
		})
	}
	ew.table(rows)
}

// renderEphemeralResources prints the ephemeral resource blocks
// (OpenTofu/Terraform 1.10+). Omitted when none exist.
func renderEphemeralResources(ew *errWriter, res []model.EphemeralResource) {
	if len(res) == 0 {
		return
	}
	ew.printf("## Ephemeral resources (%d)\n", len(res))
	rows := [][]string{{"TYPE", "NAME", "PROVIDER", "COUNT/FOR_EACH", "LIFECYCLE"}}
	for _, r := range res {
		rows = append(rows, []string{
			r.Type, r.Name, orNone(r.Provider), repetition(r.Count, r.ForEach),
			lifecycleSummary(r.Lifecycle),
		})
	}
	ew.table(rows)
}

// renderMoved prints the moved {} refactoring blocks. Omitted when none exist.
func renderMoved(ew *errWriter, moved []model.MovedBlock) {
	if len(moved) == 0 {
		return
	}
	ew.printf("## Moved (%d)\n", len(moved))
	rows := [][]string{{"FROM", "TO"}}
	for _, m := range moved {
		rows = append(rows, []string{m.From, m.To})
	}
	ew.table(rows)
}

// renderImports prints the import {} blocks. Omitted when none exist.
func renderImports(ew *errWriter, imports []model.ImportBlock) {
	if len(imports) == 0 {
		return
	}
	ew.printf("## Imports (%d)\n", len(imports))
	rows := [][]string{{"TO", "ID", "PROVIDER"}}
	for _, i := range imports {
		rows = append(rows, []string{i.To, truncate(i.ID.Source, 40), orNone(i.Provider)})
	}
	ew.table(rows)
}

// renderRemoved prints the removed {} blocks. Omitted when none exist.
func renderRemoved(ew *errWriter, removed []model.RemovedBlock) {
	if len(removed) == 0 {
		return
	}
	ew.printf("## Removed (%d)\n", len(removed))
	rows := [][]string{{"FROM", "DESTROY_ON_DROP"}}
	for _, r := range removed {
		rows = append(rows, []string{r.From, boolPtr(r.DestroyOnDrop)})
	}
	ew.table(rows)
}

// renderChecks prints the check {} blocks, summarizing each block's
// assertions as a count. Omitted when none exist.
func renderChecks(ew *errWriter, checks []model.CheckBlock) {
	if len(checks) == 0 {
		return
	}
	ew.printf("## Checks (%d)\n", len(checks))
	rows := [][]string{{"NAME", "DATA SOURCE", "ASSERTIONS"}}
	for _, c := range checks {
		data := "(none)"
		if c.DataSource != nil {
			data = c.DataSource.Type + "." + c.DataSource.Name
		}
		rows = append(rows, []string{c.Name, data, count(len(c.Assertions))})
	}
	ew.table(rows)
}

// renderEncryption prints a presence/count summary of the OpenTofu state
// and plan encryption configuration. Nil (no encryption block) is skipped.
func renderEncryption(ew *errWriter, enc *model.Encryption) {
	if enc == nil {
		return
	}
	ew.printf("## Encryption\n")
	rows := [][]string{
		{"KEY PROVIDERS", count(len(enc.KeyProviders))},
		{"METHODS", count(len(enc.Methods))},
		{"STATE", yesNo(enc.State != nil)},
		{"PLAN", yesNo(enc.Plan != nil)},
		{"REMOTE STATE SOURCES", count(len(enc.RemoteStateSources))},
	}
	ew.table(rows)
}

// renderModuleCalls prints the module {} invocations. Omitted when none exist.
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

// renderProviders prints the provider {} configuration blocks. Omitted
// when none exist.
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

// renderRequiredProviders prints the terraform.required_providers entries
// in sorted order for deterministic output. Omitted when none exist.
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

// renderDiagnostics prints the loader diagnostics with their severity and
// summary. Omitted when there are none.
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

// truncate collapses newlines to spaces and shortens s to at most n
// characters, appending an ellipsis when it has to cut.
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

// orNone returns s, or the placeholder "(none)" when s is empty.
func orNone(s string) string {
	if s == "" {
		return "(none)"
	}
	return s
}

// yesNo renders a bool as "yes" or "no".
func yesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

// count renders a slice length as a column value, using a dash for zero
// so the common "none" case reads as visual whitespace.
func count(n int) string {
	if n == 0 {
		return "-"
	}
	return strconv.Itoa(n)
}

// boolPtr renders an optional bool: unset reads as "(default)".
func boolPtr(b *bool) string {
	if b == nil {
		return "(default)"
	}
	return yesNo(*b)
}

// lifecycleSummary compactly describes a lifecycle {} block: the scalar
// flags that are set plus counts of nested pre/postcondition blocks. Full
// detail (conditions, error messages, ranges) lives in the JSON output.
func lifecycleSummary(lc *model.Lifecycle) string {
	if lc == nil {
		return "(none)"
	}
	var parts []string
	if lc.CreateBeforeDestroy != nil {
		parts = append(parts, "create_before_destroy="+yesNo(*lc.CreateBeforeDestroy))
	}
	if lc.PreventDestroy != nil {
		parts = append(parts, "prevent_destroy="+yesNo(*lc.PreventDestroy))
	}
	if len(lc.IgnoreChanges) > 0 {
		parts = append(parts, fmt.Sprintf("ignore_changes=%d", len(lc.IgnoreChanges)))
	}
	if len(lc.ReplaceTriggeredBy) > 0 {
		parts = append(parts, fmt.Sprintf("replace_triggered_by=%d", len(lc.ReplaceTriggeredBy)))
	}
	if len(lc.Preconditions) > 0 {
		parts = append(parts, fmt.Sprintf("preconditions=%d", len(lc.Preconditions)))
	}
	if len(lc.Postconditions) > 0 {
		parts = append(parts, fmt.Sprintf("postconditions=%d", len(lc.Postconditions)))
	}
	if len(parts) == 0 {
		return "(empty)"
	}
	return strings.Join(parts, "; ")
}

// repetition renders a resource's count or for_each meta-argument as a
// single column value, or "(none)" when neither is set.
func repetition(cnt, forEach *model.Expression) string {
	switch {
	case cnt != nil:
		return "count=" + truncate(cnt.Source, 20)
	case forEach != nil:
		return "for_each=" + truncate(forEach.Source, 20)
	default:
		return "(none)"
	}
}
