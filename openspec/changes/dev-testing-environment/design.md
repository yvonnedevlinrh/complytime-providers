## Context

The complyctl repository is adding a devcontainer-based testing environment
(complyctl `dev-testing-environment` change). This change mirrors that
setup in complytime-providers, following the established pattern where
complyctl owns canonical infrastructure and complytime-providers consumes
it (same as cross-repo integration tests).

The key difference from complyctl's devcontainer: here, providers are built
from the local PR branch while complyctl is cloned from main. In complyctl's
devcontainer, it's the reverse.

## Goals / Non-Goals

**Goals:**

- Provide maintainers reviewing provider PRs with a one-command path to
  a Fedora environment where they can test `complyctl` commands using
  provider binaries from the PR branch.
- Reuse complyctl's test infrastructure (mock OCI registry, Gemara test
  fixtures, workspace configuration) by cloning it at setup time.
- Maintain consistency with complyctl's devcontainer in terms of base
  image, system dependencies, and tool versions.

**Non-Goals:**

- Owning or maintaining the test content — that belongs to complyctl.
- Publishing a pre-built container image.
- Replacing the complytime-demos Vagrant setup.

## Decisions

### D1: Mirror complyctl's Containerfile

**Decision**: Use the same Containerfile structure as complyctl (Fedora 43
base, same system packages via dnf).

**Rationale**: Both devcontainers need the same system-level dependencies.
The Containerfile is ~10 lines. Keeping them identical avoids
environment discrepancies when testing.

### D2: Clone complyctl at post-create time for test fixtures

**Decision**: The post-create script clones complyctl from main to obtain
complyctl binary, mock-oci-registry binary, and test fixtures. This is
the reverse of what complyctl's devcontainer does (which clones
complytime-providers from main).

**Alternatives considered**:

- *Duplicate test fixtures into this repo*: Creates maintenance burden
  and drift risk. The cross-repo integration test already avoids this
  by running complyctl's script directly.
- *Require complyctl to be merged first and reference a release*: Adds
  fragility around release timing. Building from main is simpler and
  always current.

**Rationale**: Mirrors the cross-repo integration test pattern. The test
script, workspace config, and mock registry all live in complyctl. Cloning
at setup time ensures freshness without duplicating content.

### D3: Documentation references complyctl docs for detail

**Decision**: The providers documentation covers provider-specific setup
and references complyctl's dev-testing-environment docs for shared
concepts (Codespaces usage, DevPod setup, GITHUB_TOKEN configuration).

**Rationale**: Avoids duplicating documentation that would drift. The
complyctl docs are the canonical reference for the testing environment.

## Risks / Trade-offs

- **[Dependency on complyctl main]** → If complyctl main is broken, the
  providers devcontainer setup will fail. This is the same risk the
  cross-repo integration test already accepts. The CI workflow validates
  main continuously.

- **[Containerfile duplication]** → Accepted. The Containerfile is minimal
  and both repos need identical system packages. The alternative (shared
  image registry) was explicitly rejected to avoid complexity.

- **[Ordering dependency]** → This change must be merged after complyctl's
  `dev-testing-environment` change, since the post-create script depends
  on complyctl's test fixtures and mock-oci-registry binary existing in
  main.
