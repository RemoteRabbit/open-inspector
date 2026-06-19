# open-inspector

A modern, OpenTofu-aware inspector for Terraform / OpenTofu configurations,
usable both as a Go library and a CLI.

The config loader parses `.tf`, `.tf.json`, `.tofu`, `.tofu.json` (plus
`_override` files) into a stable, source-range-accurate `model.Module`. It
never runs `init`, never contacts providers, and for local modules needs no
network: everything is static analysis.

## Highlights

- **Config inspection** - parse a module directory into a stable,
  source-range-accurate model, capturing nested resource blocks and module
  inputs verbatim.
- **Dependency graph** - derive the intra-module resource dependency graph
  from captured references plus `depends_on`; emit JSON, DOT, Mermaid, or a
  tree view, optionally recursing into child modules with cross-module edges.
- **Module graph** - recursively resolve local, registry, git, and http
  module sources into a tree.
- **Provider schema introspection (optional)** - enrich the model with
  attribute validity and deprecation findings.
- **Modern block coverage** - `moved`, `import`, `removed`, `check`,
  `ephemeral`, OpenTofu encryption blocks, provider `for_each`,
  `configuration_aliases`, and more.

## Install

```sh
go install github.com/remoterabbit/open-inspector/cmd/open-inspector@latest
```

See [Getting started](getting-started.md) for prebuilt binaries, verification,
and full usage.

## Quick start

```sh
# Human-readable summary of a module.
open-inspector config ./path/to/module

# Machine-readable JSON (versioned schema).
open-inspector config --json ./path/to/module

# Render the dependency graph.
open-inspector graph ./path/to/module
```

## Learn more

- [Getting started](getting-started.md) - install and use the CLI.
- [CLI reference](cli/open-inspector.md) - the generated, exhaustive flag
  reference for every subcommand.
- [JSON schema (v1)](schema/v1.md) - the versioned wire format emitted by
  `--json`.
