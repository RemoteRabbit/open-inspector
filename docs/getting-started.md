# Getting started

A practical guide to installing `open-inspector` and using it from the
command line. For the generated, exhaustive flag reference see
[docs/cli](./cli); for man pages see [docs/man](./man).

## Install

### With `go install` (recommended)

If you have a matching Go toolchain (see `go.mod`):

```sh
go install github.com/remoterabbit/open-inspector/cmd/open-inspector@latest
```

The binary lands in `$(go env GOBIN)` (or `$(go env GOPATH)/bin`); make sure
that directory is on your `PATH`. Note that `go install` builds do not embed a
version string, so `--version` reports a development value. If you need an
exact, signed release, use a prebuilt binary instead.

### Prebuilt binaries

Each release ships signed archives for Linux, macOS, and Windows on both
`amd64` and `arm64`. Download the archive for your platform from the
[releases page](https://github.com/remoterabbit/open-inspector/releases),
extract it, and put the `open-inspector` binary somewhere on your `PATH`.

```sh
# Example: Linux x86_64. Replace VERSION and the asset name to match your platform.
VERSION=0.7.0
curl -sSLO "https://github.com/remoterabbit/open-inspector/releases/download/v${VERSION}/open-inspector_${VERSION}_linux_x86_64.tar.gz"
tar -xzf "open-inspector_${VERSION}_linux_x86_64.tar.gz"
sudo install open-inspector /usr/local/bin/
open-inspector --version
```

Archive names follow `open-inspector_<version>_<os>_<arch>`, where `arch` is
`x86_64` for amd64 and `arm64` for arm64; Windows builds are `.zip`.

#### Verify the download (optional)

Every release includes a `checksums.txt` plus a keyless
[cosign](https://github.com/sigstore/cosign) signature
(`checksums.txt.sigstore.json`). To confirm an archive is authentic:

```sh
# 1. Verify the checksum file was signed by this repo's release workflow.
cosign verify-blob \
  --bundle checksums.txt.sigstore.json \
  --certificate-identity-regexp 'https://github.com/remoterabbit/open-inspector/.+' \
  --certificate-oidc-issuer     https://token.actions.githubusercontent.com \
  checksums.txt

# 2. Verify your archive matches its recorded checksum.
sha256sum --check --ignore-missing checksums.txt
```

### From source

```sh
git clone https://github.com/remoterabbit/open-inspector
cd open-inspector
make build            # produces ./bin/open-inspector
./bin/open-inspector --version
```

## Usage

`open-inspector` operates on a single module directory (it reads `.tf`,
`.tf.json`, `.tofu`, and `.tofu.json` files in that directory,
non-recursively). It never runs `init`, never contacts providers, and for
local modules needs no network: everything below is static analysis.

There are three subcommands:

| Command   | What it does                                                       |
| --------- | ----------------------------------------------------------------- |
| `config`  | Inspect a module: variables, outputs, resources, providers, ...   |
| `graph`   | Render the intra-module resource dependency graph                  |
| `modules` | Render the module-call graph (which module calls which)            |

### Inspect a module (`config`)

```sh
# Human-readable summary table.
open-inspector config ./path/to/module

# Machine-readable JSON envelope (schema is versioned; see docs/schema).
open-inspector config --json ./path/to/module

# Embed the dependency graph in the output.
open-inspector config --deps ./path/to/module

# Enrich resources with provider-schema findings (attribute validity,
# deprecations). Pass a `providers schema -json` file, or "auto" to shell
# out to terraform/tofu for you.
open-inspector config --schema schema.json ./path/to/module
open-inspector config --schema auto ./path/to/module
```

### Dependency graph (`graph`)

Builds the graph of resources, data sources, locals, outputs, variables, and
module calls, with edges derived from references between them plus
`depends_on` and `replace_triggered_by`. This mirrors `terraform graph`, but
statically.

```sh
# Tree view (default).
open-inspector graph ./path/to/module

# Other formats.
open-inspector graph --format dot ./path/to/module
open-inspector graph --format mermaid ./path/to/module
open-inspector graph --format json ./path/to/module

# Recurse into child modules, drawing cross-module edges.
open-inspector graph --recursive ./path/to/module
open-inspector graph --recursive --max-depth 4 ./path/to/module

# Keep only certain node kinds.
open-inspector graph --kind resource,module ./path/to/module

# When filtering with --kind, collapse paths that ran through removed
# nodes instead of breaking them.
open-inspector graph --kind resource --contract ./path/to/module
```

Valid `--kind` values: `data`, `ephemeral`, `local`, `module`, `output`,
`resource`, `variable`.

### Module-call graph (`modules`)

Resolves local, registry, git, and http module sources into a tree showing
which module calls which.

```sh
open-inspector modules ./path/to/module                 # tree (default)
open-inspector modules --format mermaid ./path/to/module
open-inspector modules --format json ./path/to/module
```

## Exit codes

All three subcommands accept `--fail-on` to control the exit code based on
the diagnostics found while loading:

```sh
open-inspector config --fail-on error   ./module   # nonzero on any error (default)
open-inspector config --fail-on warning ./module   # nonzero on warnings too
open-inspector config --fail-on never   ./module   # always exit 0 if it ran
```

This makes the tool easy to wire into CI: a malformed module surfaces as a
nonzero exit without aborting the whole run on the first bad file.

## Global flags

These apply to every command:

| Flag                  | Purpose                                          |
| --------------------- | ------------------------------------------------ |
| `--log-level <level>` | `debug`, `info` (default), `warn`, or `error`    |
| `--quiet`             | suppress informational log output                |
| `--no-color`          | disable colored output (useful in CI/logs)       |
| `--version` / `-v`    | print the version and exit                        |

## Use as a Go library

`open-inspector` is also importable. See the
[Library section of the README](https://github.com/remoterabbit/open-inspector#library) for the `inspector.Inspect`
API and its `With*` options.
