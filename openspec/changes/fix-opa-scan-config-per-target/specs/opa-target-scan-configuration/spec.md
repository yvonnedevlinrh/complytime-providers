# Spec Delta

## Purpose

Ensure OPA bundle targets retain and use their own generated scan configuration during
single-target and multi-target compliance scans.

## ADDED Requirements

### Requirement: Generated configuration is scoped to its OPA bundle

The OPA provider SHALL retain generated requirement IDs, reverse mappings, and bundle
location separately for each distinct `opa_bundle_ref`.

#### Scenario: Generate two distinct bundle targets

- **WHEN** generation runs for two targets with different `opa_bundle_ref` values
- **THEN** each target's generated scan configuration remains available without being
  overwritten by the other target

### Requirement: Scan uses the matching generated configuration

The OPA provider SHALL use the generated scan configuration associated with each bundle target
when processing that target.

#### Scenario: Scan two generated bundle targets

- **WHEN** a scan processes two targets generated from different `opa_bundle_ref` values
- **THEN** each target uses its own requirement IDs, reverse mappings, and bundle location

#### Scenario: Reverse target order

- **WHEN** the same generated targets are scanned in a different order
- **THEN** each target produces results from the same matching configuration as before

#### Scenario: Scan one generated bundle target

- **WHEN** a scan processes one target generated from an `opa_bundle_ref`
- **THEN** the target uses its matching generated configuration as in the existing single-target
  flow
