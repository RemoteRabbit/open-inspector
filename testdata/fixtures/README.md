# Test fixtures

Each subdirectory is a self-contained Terraform / OpenTofu module used by
the loader and integration tests. Fixtures are intentionally minimal -
just enough HCL to exercise one feature or combination.

| Fixture | Status | Exercises |
|---|---|---|
| [`simple/`](./simple) | valid | minimal module: one resource, one variable, one output |
| [`variables-and-outputs/`](./variables-and-outputs) | valid | every variable/output/local feature (types, defaults, validation, sensitive, nullable, descriptions) |
| [`variable-types/`](./variable-types) | valid | structured variable types (`type_spec`) and decoded constant defaults (`default_value`): every type kind, optional object attributes, null and precision-sensitive literals |
| [`providers/`](./providers) | valid | `required_providers` with `source`, `version`, `configuration_aliases`; multiple providers; provider aliases |
| [`resources-count-foreach/`](./resources-count-foreach) | valid | `count` and `for_each` on resources, data sources, and module calls |
| [`modern-blocks/`](./modern-blocks) | valid | `moved`, `import`, `removed`, `check` blocks |
| [`ephemeral/`](./ephemeral) | valid | `ephemeral` resources, ephemeral variables and outputs (TF 1.10+ / OpenTofu 1.10+) |
| [`opentofu-encryption/`](./opentofu-encryption) | valid (OpenTofu only) | state/plan encryption block (OpenTofu 1.7+) |
| [`opentofu-provider-foreach/`](./opentofu-provider-foreach) | valid (OpenTofu only) | provider `for_each` / multi-instance providers (OpenTofu 1.9+) |
| [`tofu-extension/`](./tofu-extension) | valid (OpenTofu only) | `.tofu` and `.tofu.json` file extensions |
| [`overrides/`](./overrides) | valid | `_override.tf` and `override.tf` merge semantics |
| [`json-config/`](./json-config) | valid | `.tf.json` configuration variant |
| [`multi-module/`](./multi-module) | valid | root module calling two local child modules (graph test) |
| [`module-sources/`](./module-sources) | valid (declarations only) | registry, git, http module source declarations - not actually fetched in unit tests |
| [`invalid/syntax-error/`](./invalid/syntax-error) | invalid | unparseable HCL - diagnostic surface test |
| [`invalid/missing-required/`](./invalid/missing-required) | invalid | resource missing required arguments - diagnostic test (only meaningful with provider schema) |

## Conventions

- Fixtures are **never** initialized (`terraform init`) by tests. They
  must be parseable from source alone.
- Provider versions in `required_providers` are pinned just enough to be
  realistic; tests should not depend on resolving them.
- When a fixture is OpenTofu-only, the directory `README.md` (or a
  top-of-file comment) calls that out.
- Adding a fixture? Update the table above, add a one-line description
  inside the fixture (top-of-file comment), and write a test that
  references it by its directory name.
