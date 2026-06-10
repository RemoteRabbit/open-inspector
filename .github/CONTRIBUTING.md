# Contributing to open-inspector

Thanks for your interest! This document covers the local dev loop, the
commit-message convention, and how releases happen.

## Local setup

```sh
make pre-commit-install   # one-time: wire git hooks (commit-msg, pre-commit, pre-push)
make all                  # fmt + lint + license-check + test + build
```

Requires Go (matching `go.mod`), `golangci-lint`, and `pre-commit`. A
[`devenv`](https://devenv.sh) shell, plus `.tool-versions` for
[`asdf`](https://asdf-vm.com) / [`mise`](https://mise.jdx.dev), are
provided so versions match CI.

## Pre-commit hooks

`.pre-commit-config.yaml` runs three sets of checks:

- **commit-msg**: Conventional Commit format (see below).
- **pre-commit**: file hygiene, `gofmt -s`, `go vet`,
  `golangci-lint --fix`, MPL license header, `go mod tidy` (when
  `go.mod`/`go.sum` change).
- **pre-push**: `go test -race`.

Every hook shells out to the same `make` target CI uses, so there is
one source of truth per check.

## Commit message convention

We use [Conventional Commits](https://www.conventionalcommits.org/).
The `commit-msg` hook enforces the format; release-please consumes the
type to decide the next version and the CHANGELOG section.

**Format:** `type: description` or `type(scope): description`
(scope is optional; type must be lowercase).

| Type       | CHANGELOG section          | Version bump        | Visible? |
|------------|----------------------------|---------------------|----------|
| `feat`     | Features                   | minor               | ✅       |
| `fix`      | Bug Fixes                  | patch               | ✅       |
| `perf`     | Performance Improvements   | patch               | ✅       |
| `revert`   | Reverts                    | patch               | ✅       |
| `deps`     | Dependencies               | patch               | ✅       |
| `docs`     | Documentation              | patch               | ✅       |
| `refactor` | Code Refactoring           | patch               | hidden   |
| `test`     | Tests                      | patch               | hidden   |
| `build`    | Build System               | patch               | hidden   |
| `ci`       | Continuous Integration     | patch               | hidden   |
| `chore`    | Miscellaneous Chores       | patch               | hidden   |
| `style`    | Styles                     | patch               | hidden   |

Pre-1.0 note: `feat` currently bumps the **minor** version
(`bump-minor-pre-major: true` in `release-please-config.json`).
After we tag `v1.0.0`, `feat` will start bumping the minor as usual
and only `feat!` / `BREAKING CHANGE` will bump major.

### Breaking changes

Append `!` after the type, or include a `BREAKING CHANGE:` footer.
Either triggers a **major** bump:

```text
feat!: drop support for Terraform 0.12

BREAKING CHANGE: the loader now requires HCL2 input; pre-0.12
configurations will fail with a clear diagnostic instead of being
partially parsed.
```

### Examples

```text
feat: parse provider for_each blocks
feat(graph): emit Mermaid output
fix: handle CRLF line endings in .tf.json files
perf: cache HCL file body across overrides merge
docs: clarify OpenTofu encryption block support
ci: pin actions to commit SHAs
chore: bump devenv pin
```

Renovate writes commits as `fix(deps):` for runtime Go modules,
`chore(deps):` for indirect/dev-tool bumps, and `ci(deps):` for
GitHub Actions, all already permitted.

## Release flow

You do not tag manually. The pipeline is:

1. Merge a PR into `main` with a Conventional Commit title.
2. The `release-please` workflow opens or updates a **Release PR**
   containing the next version bump, `CHANGELOG.md` entries, and the
   updated `Version` constant in `pkg/inspector/inspector.go`.
3. When ready to ship, merge that Release PR. It tags `vX.Y.Z`.
4. The `release` workflow runs GoReleaser on the tag: cross-compiled
   binaries (linux/darwin/windows × amd64/arm64), SBOMs, a
   cosign-signed `checksums.txt`, and a GitHub Release with the
   release-please-generated notes.

## Reporting bugs and proposing features

Open a GitHub issue. For non-trivial features, please describe the
intended user-facing behavior before opening a PR so we can agree on
shape and scope.

## License

By contributing, you agree your contributions are licensed under
[MPL-2.0](./LICENSE), matching the rest of the project.
