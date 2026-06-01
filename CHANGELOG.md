# Changelog

## Unreleased

### Features

- **opa-provider**: Implement Generate RPC with mapping-based filtering. Generate reads `complytime-mapping.json` from the OCI policy bundle, matches Gemara assessment plan RequirementIDs to Rego namespaces, and writes a `scan-config.json` artifact. Scan uses the artifact for namespace-filtered conftest evaluation (`--namespace` flags) and resolves Rego-derived IDs to Gemara RequirementIDs in results. Bundles without a mapping file fall back to `--all-namespaces` with Rego-derived IDs.
