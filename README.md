# open-inspector

A modern, OpenTofu-aware inspector for Terraform / OpenTofu configurations,
usable both as a Go library and a CLI.

> **Status:** the config loader is complete: it parses every
> Terraform / OpenTofu module directory listed under "Compatibility"
> below into a stable, source-range-accurate `model.Module`. The CLI is
> still a thin scaffold; module-graph resolution and provider-schema
> enrichment are the next milestones.

## Goals

- **Config inspection** - parse `.tf`, `.tf.json`, `.tofu`, `.tofu.json`
  (plus `_override` files) into a stable, source-range-accurate model.
- **Module graph** - recursively resolve local, registry, git, and http
  module sources; emit JSON, DOT, Mermaid, or a tree view.
- **Provider schema introspection (optional)** - enrich the model with
  attribute validity and deprecation findings from
  `terraform providers schema -json` or its `tofu` equivalent.
- **Modern block coverage** - `moved`, `import`, `removed`, `check`,
  `ephemeral`, OpenTofu encryption blocks, provider `for_each`,
  `configuration_aliases`, and other features the original
  `terraform-config-inspect` never learned.

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
  alongside their `.tf` cousins. OpenTofu-specific blocks like
  `encryption {}` and provider `for_each` parse without errors today but
  their fields are not yet surfaced (see "Coming next" below).
- **Forward compatibility** is automatic: when Terraform or OpenTofu
  adds a new block or attribute, the loader will load the file cleanly
  and ignore the new construct until a future release of open-inspector
  adds explicit decoding for it.

The legacy pre-0.13 shorthand `required_providers { aws = "~> 4.0" }`
is still accepted. Pre-0.12 HCL (the untyped-attribute syntax) is not
supported.

### Recognized but not yet decoded

The following constructs parse cleanly (no diagnostics, no panics) but
their model fields are unpopulated for now:

- `moved {}` (TF 1.1+), `import {}` (TF 1.5+), `removed {}` (TF 1.7+),
  `check {}` (TF 1.5+)
- `ephemeral "type" "name" {}` resource blocks (TF 1.10+ / OpenTofu 1.10+)
- OpenTofu `terraform { encryption {} }` (OpenTofu 1.7+)
- OpenTofu provider `for_each` and multi-instance providers (OpenTofu 1.9+)
- Override files (`_override.tf`, `override.tf`): collected by the
  walker but not yet merged into the base files

## Coming next

Near-term:

- **modern block coverage:** decode every construct listed
  above, plus override file merging per Terraform's documented
  last-wins semantics.
- **CLI:** replace the flag scaffold with cobra; add
  `open-inspector config <dir>` with a human table renderer and
  versioned JSON output; configurable exit codes via
  `--fail-on=error|warning|never`; auto-generated markdown and man-page
  docs.
- **module graph:** recursively resolve `module` calls across
  local, registry, git, and HTTP sources; emit JSON, DOT, Mermaid, or
  tree views; cache downloaded modules under `$XDG_CACHE_HOME`.
- **provider schema enrichment, optional:** annotate the model
  with unknown attributes, deprecations, and missing-required findings
  from `tofu providers schema -json` output.

## Usage (today)

```sh
make build
./bin/open-inspector .
./bin/open-inspector --json .
./bin/open-inspector --version
```

## Library

```go
import "github.com/remoterabbit/open-inspector/pkg/inspector"

mod, err := inspector.Inspect("./path/to/module")
```

## Development

```sh
make pre-commit-install                  # one-time: wire git hooks (commit + push)
make all                                 # fmt + lint + license-check + test + build
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

## License

[MPL-2.0](./LICENSE). Matches the OpenTofu / HCL ecosystem: derivative
work to the library files stays MPL, but the library can be embedded in
projects under any license.
