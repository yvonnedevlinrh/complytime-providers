## 1. Containerfile

- [ ] 1.1 Create `.devcontainer/Containerfile` using
  `registry.fedoraproject.org/fedora:43` as base image with dnf install of
  `openscap-scanner`, `scap-security-guide`, `golang`, `git`, and `make`
  (mirror complyctl's Containerfile)

## 2. Post-create setup script

- [ ] 2.1 Create `.devcontainer/scripts/post-create.sh` that builds
  both provider binaries from local source (`make build`)
- [ ] 2.2 Add `go install` of snappy
  (`github.com/carabiner-dev/snappy@latest`) and ampel
  (`github.com/carabiner-dev/ampel/cmd/ampel@latest`) to the script
- [ ] 2.3 Add clone of `complyctl@main`, build complyctl and
  mock-oci-registry, and add complyctl binary to PATH
- [ ] 2.4 Copy provider binaries (`complyctl-provider-openscap` and
  `complyctl-provider-ampel`) to `~/.complytime/providers/`
- [ ] 2.5 Add workspace setup: copy `complytime.yaml` and granular policies
  from the cloned complyctl's `tests/cross-repo/testdata/` to a test
  workspace directory
- [ ] 2.6 Add mock OCI registry startup in background with readiness check
  (mirror the pattern from complyctl's
  `tests/cross-repo/cross_repo_integration_test.sh`)

## 3. Devcontainer configuration

- [ ] 3.1 Create `.devcontainer/devcontainer.json` referencing the
  Containerfile and setting `postCreateCommand` to run the post-create
  script

## 4. Documentation

- [ ] 4.1 Create `docs/dev-testing-environment.md` covering
  provider-specific setup and referencing complyctl's documentation
  for shared concepts (Codespaces usage, DevPod setup, GITHUB_TOKEN
  configuration)
- [ ] 4.2 Update `README.md` to add a link to the dev testing environment
  documentation

## 5. Verification

- [ ] 5.1 Test the devcontainer locally with `podman` or `docker` to verify
  the Containerfile builds, the post-create script completes, and
  `complyctl get` and `complyctl generate --policy-id test-ampel-bp`
  succeed using provider binaries built from local source
