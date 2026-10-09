# Design

## Context

See `proposal.md` for motivation. `Generate` currently writes one configuration under
`opa/generated`, and `Scan` reads it once before processing all targets. OPA bundle content is
already stored in a deterministic per-bundle directory. Scan requests do not carry the
`ComplypackContentPath` supplied to Generate.

## Goals / Non-Goals

**Goals:**

- Keep each OPA bundle's generated configuration with its existing bundle directory.
- Load configuration in the target-processing path so namespace selection, requirement mapping,
  and passing-assessment synthesis use the matching target state.
- Preserve current response grouping and single-target behavior.

**Non-Goals:**

- Target-scoping complypack configuration, which requires separate request-path design.
- Changing AMPEL; issue #194 owns its distinct storage strategy.
- Adding overwrite warnings, concurrency hardening, a provider-wide abstraction, or new artifact
  formats.

## Decisions

### Store bundle configuration in the existing bundle directory

For `opa_bundle_ref` generation, pass the resolved policy directory to the existing scan-config
writer. This reuses `PolicyDirForBundle` and keeps the configuration beside the exact policy it
describes.

The alternative was a bundle-keyed hierarchy under `opa/generated`. It was rejected because it
duplicates bundle-key path logic and adds another directory relationship without improving the
required behavior.

### Resolve configuration per target during Scan

Move configuration lookup from the start of `Scan` into per-target processing. A target with an
`opa_bundle_ref` first uses the configuration in that bundle's policy directory. The existing
shared generated directory remains only as the complypack-compatible fallback because Scan has no
complypack path from which to derive a target-specific location.

Each target's configuration must remain associated with its results through response conversion.
The implementation should adapt the existing processing and response assembly directly rather
than introduce a generalized configuration registry or cross-provider abstraction.

### Reuse existing scan-config format and bundle cache

`WriteScanConfig` and `ReadScanConfig` already accept arbitrary directories, so their format and
API remain unchanged. Existing bundle caching remains responsible only for avoiding repeated
pulls; it does not become a second configuration store.

## Risks / Trade-offs

- [Shared complypack fallback remains target-agnostic] -> Keep it unchanged and defer target
  scoping until Scan receives a stable complypack identifier.
- [Per-target mapping can accidentally be flattened during response assembly] -> Add one focused
  end-to-end regression with different namespaces and requirement mappings for two targets.
- [Existing workspaces may contain only the former shared bundle configuration] -> A fresh
  Generate recreates configuration in the bundle directory; no migration artifact is added.

## Migration Plan

No persistent migration is required. Running Generate writes bundle configurations to their new
locations. Rollback restores the former shared-file read and write behavior.
