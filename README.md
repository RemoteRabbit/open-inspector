# open-inspector

A modern, OpenTofu-aware inspector for Terraform / OpenTofu configurations,
usable both as a Go library and a CLI.

> **Status:** early scaffold. Today it only proves the build/test/lint loop.
> Real config parsing, module-graph resolution, and provider-schema
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
