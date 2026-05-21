## ADDED Requirements

### Requirement: Devcontainer configuration provides a Fedora testing environment

The repository SHALL include a `.devcontainer/` directory with a
`devcontainer.json` and `Containerfile` that together define a Fedora-based
testing environment. The Containerfile SHALL use
`registry.fedoraproject.org/fedora:43` (or later) as the base image and
SHALL install `openscap-scanner`, `scap-security-guide`, `golang`, `git`,
and `make` via dnf.

#### Scenario: Devcontainer configuration is present and valid

- **WHEN** a maintainer opens the repository in a devcontainer-compatible
  tool (GitHub Codespaces, DevPod, or VS Code Dev Containers)
- **THEN** the tool SHALL detect `.devcontainer/devcontainer.json` and
  build the environment from the Containerfile

#### Scenario: Container is based on Fedora with provider dependencies

- **WHEN** the devcontainer finishes building
- **THEN** the running environment SHALL be Fedora-based with
  `openscap-scanner` and `scap-security-guide` available as installed
  packages

### Requirement: Post-create setup builds providers from PR and complyctl from main

The devcontainer SHALL define a post-create command (script) that builds
provider binaries from the local source (the PR branch), clones and builds
complyctl and mock-oci-registry from main, installs snappy and ampel via
`go install`, copies provider binaries to the discovery path
(`~/.complytime/providers/`), configures a test workspace using complyctl's
test fixtures, and starts the mock OCI registry in the background.

#### Scenario: Provider binaries are built from local source

- **WHEN** the devcontainer post-create command completes
- **THEN** `complyctl-provider-openscap` and `complyctl-provider-ampel`
  SHALL be built from the local repository source and present in
  `~/.complytime/providers/`

#### Scenario: complyctl is built from main

- **WHEN** the devcontainer post-create command completes
- **THEN** `complyctl` and `mock-oci-registry` SHALL be built from the
  latest complyctl main branch and available on the PATH or in expected
  binary directories

#### Scenario: External tools are installed

- **WHEN** the devcontainer post-create command completes
- **THEN** `snappy` and `ampel` SHALL be available on the PATH

#### Scenario: Test workspace is configured with Gemara content

- **WHEN** the devcontainer post-create command completes
- **THEN** a `complytime.yaml` workspace configuration SHALL exist pointing
  to the mock OCI registry, and the mock registry SHALL be running and
  serving Gemara catalogs and policies on localhost

#### Scenario: complyctl commands work end-to-end with PR providers

- **WHEN** a maintainer runs `complyctl get` followed by
  `complyctl generate --policy-id test-ampel-bp` in the test workspace
- **THEN** both commands SHALL complete successfully using the provider
  binaries built from the local PR branch

### Requirement: Documentation explains provider testing workflows

The repository SHALL include documentation in `docs/` that explains how
to use the devcontainer for PR review testing of provider changes. The
documentation SHALL reference complyctl's dev-testing-environment
documentation for shared concepts and SHALL cover provider-specific
setup details.

#### Scenario: Documentation is discoverable from README

- **WHEN** a maintainer reads the repository README
- **THEN** they SHALL find a link or reference to the dev testing
  environment documentation

#### Scenario: Documentation references complyctl docs

- **WHEN** a maintainer reads the dev testing environment documentation
- **THEN** they SHALL find references to complyctl's documentation for
  Codespaces usage, DevPod setup, and GITHUB_TOKEN configuration
