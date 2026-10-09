# Proposal

## Why

The OPA provider stores generated scan configuration in one shared file, so generating
multiple targets with different bundles overwrites earlier target state. A later scan can
therefore apply the wrong requirement IDs, reverse mapping, and bundle directory and report
incorrect compliance results.

## What Changes

- Store generated scan configuration for an `opa_bundle_ref` in that bundle's existing policy
  directory.
- Resolve and load scan configuration separately for each target during a multi-target scan.
- Preserve the existing shared configuration behavior for complypack inputs, which cannot yet
  be associated with a target from `ScanRequest`.
- Add focused regression coverage for two generated bundle targets followed by one multi-target
  scan.

## Capabilities

### New Capabilities

- `opa-target-scan-configuration`: Defines target-scoped generated scan configuration for OPA
  bundle targets.

### Modified Capabilities

None.

## Impact

- Affects OPA provider configuration persistence and per-target scan orchestration.
- Reuses existing bundle directories and scan-config read/write functions; no API or dependency
  changes are required.
- Does not change AMPEL behavior, introduce overwrite warnings or concurrency hardening, or solve
  target scoping for `ComplypackContentPath`.
