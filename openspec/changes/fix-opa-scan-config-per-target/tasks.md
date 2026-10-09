# Tasks

## 1. Target-Scoped OPA Configuration

- [ ] 1.1 Add a focused server regression using existing mocks that generates two targets with
  different bundle references, scans them in both orders, and verifies each target's namespaces
  and requirement mapping remain isolated.
- [ ] 1.2 Write bundle-generated scan configuration to the resolved bundle directory while keeping
  the current shared location for complypack generation; verify the focused Generate tests pass.
- [ ] 1.3 Resolve scan configuration per target and preserve target-specific mapping through response
  assembly without adding a new package or generalized abstraction; verify the regression and
  existing OPA server tests pass.

## 2. Integration Verification

- [ ] 2.1 Run `make build` and `make test` to reproduce the repository CI build-and-test gate.
- [ ] 2.2 Run `make lint` and confirm documentation impact is limited to the change artifacts, with no
  README, AGENTS, or changelog update required for this internal bug fix.
