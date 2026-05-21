## Why

Maintainers reviewing provider PRs need to test CLI UX changes by running
`complyctl` commands with the provider binaries from the PR branch. Today
this requires a complex multi-repository setup via complytime-demos with
Vagrant and Ansible. The complyctl repository is adding a devcontainer-based
testing environment (see complyctl `dev-testing-environment` change); this
change adds the corresponding devcontainer to complytime-providers so
maintainers can test provider PRs with the same one-command workflow.

This change is intended to be merged after the complyctl devcontainer
change, following the established pattern where complyctl owns canonical
infrastructure and complytime-providers consumes it.

## What Changes

- Add a `.devcontainer/` configuration to complytime-providers providing
  a Fedora-based testing environment with all dependencies pre-installed.
- The devcontainer builds provider binaries from the local PR branch,
  clones and builds complyctl from main, installs snappy and ampel via
  `go install`, and configures a test workspace using complyctl's existing
  test fixtures (mock OCI registry, Gemara content, workspace config).
- The Containerfile mirrors the one in complyctl (Fedora 43 + system deps).
  The post-create script is adapted to reverse the build order: providers
  from local source, complyctl from main.
- Add documentation in `docs/` with a reference in `README.md`.

## Capabilities

### New Capabilities

- `dev-testing-environment`: Devcontainer configuration and setup automation
  providing a one-command Fedora environment for interactive provider
  testing during PR reviews.

### Modified Capabilities

(none)

## Impact

- **New files**: `.devcontainer/Containerfile`, `.devcontainer/devcontainer.json`,
  `.devcontainer/scripts/post-create.sh`, `docs/dev-testing-environment.md`
- **Modified files**: `README.md` (add link to new docs)
- **Dependencies**: Uses `registry.fedoraproject.org/fedora:43` as the base
  container image. Installs `openscap-scanner`, `scap-security-guide` via dnf.
  Installs `snappy` and `ampel` via `go install` from `carabiner-dev`. Clones
  `complyctl` from GitHub at build time.
- **No changes to existing code**: Purely additive infrastructure.
- **Upstream dependency**: Requires the complyctl `dev-testing-environment`
  change to be merged first, as this change reuses complyctl's test fixtures
  and mock OCI registry.
