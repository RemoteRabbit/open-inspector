# open-inspector

A modern, OpenTofu-aware inspector for Terraform / OpenTofu configurations,
usable both as a Go library and a CLI.

> **Status:** the config loader, the cobra CLI (`config`, `graph`, and
> `modules` subcommands), module-call resolution (local, registry, git, and
> http sources), the intra-module dependency graph, and optional
> provider-schema enrichment are all in place. See "Coming next" below.

## Goals

- **Config inspection** - parse `.tf`, `.tf.json`, `.tofu`, `.tofu.json`
  (plus `_override` files) into a stable, source-range-accurate model,
  capturing nested resource blocks and module inputs verbatim.
- **Dependency graph** - derive the resource dependency graph of a module
  (resources, data sources, locals, outputs, variables, module calls) from
  captured references; emit JSON, DOT, Mermaid, or a tree view, optionally
  recursing into child modules with cross-module edges.
- **Module graph** - recursively resolve local, registry, git, and http
  module sources into a tree; emit JSON, DOT, Mermaid, or a tree view.
- **Provider schema introspection (optional)** - enrich the model with
  attribute validity and deprecation findings from
  `terraform providers schema -json` or its `tofu` equivalent.
- **Modern block coverage** - `moved`, `import`, `removed`, `check`,
  `ephemeral`, OpenTofu encryption blocks, provider `for_each`,
  `configuration_aliases`, and other features the original
  `terraform-config-inspect` never learned.

## Design thesis

A few principles drive every decision in this codebase. They explain why the
model looks the way it does and what the project deliberately refuses to do.

- **The model is the product.** The deliverable is a stable,
  JSON-serializable, source-range-accurate `model.Module`, consumed
  identically by the Go library, the CLI, and downstream tools. `pkg/model`
  and `pkg/inspector` are the public contract; lower-level packages
  (`config`, `graph`, `schema`) may change without notice until v1.
- **Capture, never evaluate.** Every expression is stored verbatim (its
  source bytes plus its range) and is never resolved by the loader.
  Downstream consumers decide whether and how to evaluate. This preserves
  authoring detail that evaluation would erase, such as the
  `optional(T, default)` markers in a variable type.
- **Partial, forward-compatible decoding.** Decoding is schema-driven:
  blocks and attributes the loader does not yet understand are ignored, not
  rejected. A module that uses a brand-new Terraform or OpenTofu feature
  still loads cleanly; explicit support can be added later without breaking
  older files.
- **Diagnostics, not failures.** Malformed configuration produces
  `Diagnostics` on the result rather than aborting. Only system-level
  problems (for example, an unreadable directory) return a Go `error`. One
  bad file never sinks the whole inspection.
- **Deterministic, byte-identical output.** Paths are normalized to forward
  slashes, `\r\n` is collapsed to `\n`, maps and locals are sorted, and
  slices keep encounter order. Output is identical across Linux, macOS, and
  Windows, and golden snapshots enforce it.
- **OpenTofu is a first-class peer.** `.tofu` / `.tofu.json` files,
  `encryption {}` blocks, and provider `for_each` are handled alongside
  their Terraform equivalents, not bolted on as an afterthought.
- **The schema is a versioned contract.** `model.SchemaVersion` and the
  generated JSON Schema (`docs/schema/`) define the wire format. Changes are
  additive (new `omitempty` fields) unless a deliberate, breaking version
  bump is made. While the project is pre-1.0 with no downstream consumers,
  breaking model changes may still ship without a bump; this tightens once
  the wire format has consumers.

### Hard decisions and trade-offs

- **Verbatim over typed (for now).** Variable types and defaults are kept as
  source strings, not decoded trees. This chose round-trip fidelity over
  consumer convenience; richer structured forms are additive future work.
- **HCL2 floor (Terraform 0.12+).** The pre-0.12 untyped-attribute syntax is
  unsupported. The legacy `required_providers { aws = "~> 4.0" }` shorthand
  is still accepted.
- **Best-effort over strict.** `NestedBody` captures resource attributes
  and nested blocks for native HCL only (JSON bodies are left unwalked), and
  override merging silently ignores override blocks with no matching target
  (tracked as a TODO). Loading something useful beats rejecting the input.
- **Out-of-process, optional schema enrichment.** Provider details come from
  `terraform`/`tofu providers schema -json`; the project never executes
  providers itself, and enrichment failures degrade to warnings rather than
  errors.
- **One source of truth for builds.** CI and the git hooks both shell out to
  the same `make` targets, so local and remote checks can never drift.

## Compatibility

The config loader uses schema-driven partial decoding, which means it
ignores blocks and attributes it does not yet understand instead of
erroring on them. The practical compatibility floor is:

- **Terraform** 0.12 or newer (when the modern HCL2 syntax landed).
  Every major feature added since then is recognized: object-form
  `required_providers` (0.13+), `for_each` (0.13+), `validation` blocks
  (0.13+), `configuration_aliases` (0.15+), `precondition` /
  `postcondition` / `replace_triggered_by` (1.2+), `nullable` (1.1+),
  `optional(T, default)` in type expressions (1.3+, preserved verbatim).
- **OpenTofu** all versions. `.tofu` and `.tofu.json` files are walked
  alongside their `.tf` cousins, and OpenTofu-specific constructs like the
  `encryption {}` block and provider `for_each` are decoded into the model.
- **Forward compatibility** is automatic: when Terraform or OpenTofu
  adds a new block or attribute, the loader will load the file cleanly
  and ignore the new construct until a future release of open-inspector
  adds explicit decoding for it.

The legacy pre-0.13 shorthand `required_providers { aws = "~> 4.0" }`
is still accepted. Pre-0.12 HCL (the untyped-attribute syntax) is not
supported.

## Coming next

The loader, CLI, both graphs, and schema enrichment are in place, along with
reference extraction, doc-comment capture, structured variable types, nested
resource bodies, and module-input capture. Remaining work is mostly about
closing known gaps:

- **JSON-body nested capture:** `NestedBody` walks native HCL only; nested
  blocks in `.tf.json` / `.tofu.json` resources are not yet captured.
- **Static policy / lint surface:** the dependency graph plus nested bodies
  are the building blocks for static checks (unused variables, unreferenced
  outputs, "what touches what" impact analysis) without evaluating config.

## Usage

```sh
make build

# Inspect a module directory.
./bin/open-inspector config ./path/to/module
./bin/open-inspector config --json ./path/to/module
./bin/open-inspector config --schema schema.json ./path/to/module

# Render the dependency graph (resources, locals, outputs, ...).
./bin/open-inspector graph ./path/to/module                  # tree (default)
./bin/open-inspector graph --format dot ./path/to/module     # also mermaid|json
./bin/open-inspector graph --recursive ./path/to/module      # include child modules
./bin/open-inspector graph --kind resource,module ./path/to/module
./bin/open-inspector graph --kind resource --contract ./path/to/module  # collapse paths through filtered-out nodes

# Render the module call graph (which module calls which).
./bin/open-inspector modules ./path/to/module                # tree (default)
./bin/open-inspector modules --format mermaid ./path/to/module

./bin/open-inspector --version
```

The dependency graph (`graph`) is a static analysis: no providers, no
`init`, and (for local modules) no network are required. All subcommands
accept `--fail-on=error|warning|never` to control the exit code. See
[docs/cli](./docs/cli) for the generated command reference and
[docs/man](./docs/man) for man pages.

## Library

```go
import "github.com/remoterabbit/open-inspector/pkg/inspector"

// Parse the module into a stable model.Module.
mod, err := inspector.Inspect("./path/to/module")

// Opt into extras: the dependency graph, recursive module resolution, and
// provider-schema enrichment.
mod, err = inspector.Inspect("./path/to/module",
    inspector.WithDependencyGraph(),
    inspector.WithModuleGraph(),
    inspector.WithSchemaAuto(),
)
```

## Development

```sh
make pre-commit-install                  # one-time: wire git hooks (commit + push)
make all                                 # fmt + lint + spell + license-fix + test + build + docs-cli + jsonschema
make license                             # add the MPL header to new source files
make pre-commit                          # run every pre-commit hook against all files
```

Requires Go (matching `go.mod`), `golangci-lint`, and `pre-commit`. A
[`devenv`](https://devenv.sh) shell is provided. License header
enforcement uses [`addlicense`](https://github.com/google/addlicense)
via `go run`; no local install needed.

[`asdf`](https://asdf-vm.com) and [`mise`](https://mise.jdx.dev) users
can run `asdf install` / `mise install` to pick up the versions pinned
in [`.tool-versions`](./.tool-versions).

### Pre-commit hooks

`.pre-commit-config.yaml` runs on every commit:

- File hygiene: trailing whitespace, EOF newlines, LF endings, merge
  conflicts, large-file guard, case conflicts, YAML/JSON validity,
  broken symlinks, private-key detection.
- Go: `gofmt -s`, `go vet`, `golangci-lint --fix`, MPL license header
  check, `go mod tidy` (when `go.mod` / `go.sum` change).

On `git push` it additionally runs `go test -race`. Every hook shells
out to the same `make` targets CI uses - single source of truth.

See [CONTRIBUTING.md](./.github/CONTRIBUTING.md) for the commit-message
convention and the release flow.

## License

[MPL-2.0](./LICENSE). Matches the OpenTofu / HCL ecosystem: derivative
work to the library files stays MPL, but the library can be embedded in
projects under any license.
