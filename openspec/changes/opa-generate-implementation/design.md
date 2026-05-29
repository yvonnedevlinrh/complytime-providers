## Architecture

Follow the AMPEL provider's declarative mapping pattern. The mapping between Gemara RequirementIDs and Rego namespaces is declared in a sidecar JSON file shipped inside the OCI policy bundle, owned by the policy author.

```
                complyctl generate --policy-id k8s-policy
                              │
                    complyctl routes by executor.id: "opa"
                              │
                              ▼
                    OPA Provider Generate RPC
                    ┌─────────────────────────────┐
                    │ 1. Validate req.Configuration│
                    │ 2. Check tools (conftest,git)│
                    │ 3. Ensure directories        │
                    │ 4. Pull OCI bundle            │
                    │ 5. Read complytime-mapping   │
                    │    .json from bundle dir      │
                    │ 6. Match RequirementIDs       │
                    │    (exact map lookup)         │
                    │ 7. Write scan-config.json     │
                    └──────────────┬──────────────┘
                                   │
                                   ▼
                    .complytime/opa/generated/
                      scan-config.json
```

## Mapping File Format

Policy bundle authors include `complytime-mapping.json` in their OCI bundle:

```json
{
  "version": "1",
  "mappings": [
    {
      "requirement_id": "CIS-K8S-5.2.6",
      "namespace": "kubernetes.run_as_root"
    },
    {
      "requirement_id": "CIS-K8S-5.4.1",
      "namespace": "kubernetes.resource_limits"
    }
  ]
}
```

- `requirement_id`: Must match the Gemara assessment plan's `RequirementID` exactly (like AMPEL's direct map lookup).
- `namespace`: The Rego `package` name that conftest recognizes for `--namespace` filtering.
- `version`: Schema version for forward compatibility.

### Why This Format

The AMPEL provider established the precedent: each granular policy file declares its own `"id"` field, and `MatchPolicies` does exact map lookup (`granular[reqID]`). The OPA mapping file follows the same principle -- policy authors explicitly declare which Rego namespace corresponds to which compliance requirement. This avoids the fragility of deriving IDs from Rego naming conventions.

The OpenSCAP approach (prefix concatenation with XCCDF rule naming) is not suitable because Rego has no ecosystem-wide naming standard like ComplianceAsCode/SSG provides for XCCDF.

## Generation Artifact Format

Generate writes `.complytime/opa/generated/scan-config.json`:

```json
{
  "namespaces": ["kubernetes.run_as_root", "kubernetes.resource_limits"],
  "reverse_mapping": {
    "kubernetes.run_as_root": "CIS-K8S-5.2.6",
    "kubernetes.resource_limits": "CIS-K8S-5.4.1"
  },
  "bundle_dir": ".complytime/opa/policy/ghcr.io_myorg_opa-policies_v1.0",
  "generated_at": "2026-05-29T14:30:00Z"
}
```

- `namespaces`: List of Rego namespaces to pass as `--namespace` flags. `null` means use `--all-namespaces` (fallback).
- `reverse_mapping`: Maps Rego-derived IDs (output of `deriveIDFromQuery`) back to Gemara RequirementIDs. `null` means use `deriveIDFromQuery` as-is (fallback).
- `bundle_dir`: Path to the pulled bundle for Scan to reference.
- `generated_at`: ISO 8601 timestamp for freshness tracking.

## Generate Flow

```go
func (s *ProviderServer) Generate(ctx context.Context, req *provider.GenerateRequest) (*provider.GenerateResponse, error) {
    // 1. Validate
    if len(req.Configuration) == 0 {
        return &provider.GenerateResponse{
            Success: false, ErrorMessage: "no assessment configurations provided",
        }, nil
    }

    // 2. Check tools
    missing, err := s.opts.ToolChecker()
    // return error if tools missing

    // 3. Ensure directories (including new GeneratedDirPath)
    cfg := config.NewConfig(s.opts.WorkspaceDir)
    cfg.EnsureDirectories()

    // 4. Pull bundle
    vars := mergeVariables(req.GlobalVariables, req.TargetVariables)
    bundleRef := vars["opa_bundle_ref"]
    if bundleRef == "" {
        return &provider.GenerateResponse{
            Success: false, ErrorMessage: "opa_bundle_ref variable is required",
        }, nil
    }
    policyDir := cfg.PolicyDirForBundle(bundleRef)
    scan.PullBundle(bundleRef, policyDir, s.opts.Runner)

    // 5. Read mapping file
    mapping, err := generate.LoadMapping(policyDir)
    if err != nil {
        logger.Warn("no complytime-mapping.json found, skipping requirement filtering")
        generate.WriteScanConfig(cfg.GeneratedDirPath(), nil, nil, policyDir)
        return &provider.GenerateResponse{Success: true}, nil
    }

    // 6. Match RequirementIDs (exact lookup, like AMPEL)
    namespaces, reverseMap, warnings := generate.MatchRequirements(req.Configuration, mapping)
    for _, w := range warnings {
        logger.Warn(w)
    }

    // 7. Write scan-config.json
    generate.WriteScanConfig(cfg.GeneratedDirPath(), namespaces, reverseMap, policyDir)

    return &provider.GenerateResponse{Success: true}, nil
}
```

## Scan Changes

Scan reads `scan-config.json` before evaluating. The change is localized to `processTarget`:

```go
// In processTarget, after bundle resolution:
scanCfg, err := generate.ReadScanConfig(cfg.GeneratedDirPath())
if err == nil && scanCfg.Namespaces != nil {
    raw, err = scan.EvalPolicyWithNamespaces(inputPath, policyDir, scanCfg.Namespaces, s.opts.Runner)
} else {
    raw, err = scan.EvalPolicy(inputPath, policyDir, s.opts.Runner)
}
```

New function in `scan/scan.go`:

```go
func constructConftestTestCommandWithNamespaces(inputPath, policyDir string, namespaces []string) (string, []string) {
    args := []string{"test", inputPath, "--policy", policyDir, "--output", "json", "--no-fail"}
    for _, ns := range namespaces {
        args = append(args, "--namespace", ns)
    }
    return "conftest", args
}
```

## Results Changes

`ToScanResponse` accepts an optional reverse mapping:

```go
func ToScanResponse(targetResults []*PerTargetResult, reverseMap map[string]string) *provider.ScanResponse {
    // ... existing grouping logic ...
    for _, f := range tr.Findings {
        reqID := f.RequirementID
        if reverseMap != nil {
            if gemaraID, ok := reverseMap[reqID]; ok {
                reqID = gemaraID
            }
        }
        // group by reqID ...
    }
}
```

The reverse mapping is passed from `processTarget` through `evalAndParse` to `ToScanResponse`. When the scan config is absent or has no mapping, `nil` is passed and behavior is identical to today.

## Fallback Behavior

| Bundle has mapping? | Generate writes | Scan uses | Result IDs |
|---------------------|----------------|-----------|------------|
| Yes | `namespaces: [...]`, `reverse_mapping: {...}` | `--namespace ns1 --namespace ns2` | Gemara IDs |
| No | `namespaces: null`, `reverse_mapping: null` | `--all-namespaces` | Rego-derived |

Both modes are first-class. The fallback is permanent, not deprecated.

## Workspace Layout After Generate

```
.complytime/opa/
  policy/
    ghcr.io_myorg_opa-policies_v1.0/
      kubernetes/run_as_root.rego
      kubernetes/resource_limits.rego
      docker/network_encryption.rego
      complytime-mapping.json          ← read by Generate
  generated/                           ← NEW directory
    scan-config.json                   ← written by Generate, read by Scan
  repos/
  results/
```

## Dependencies

- No new Go module dependencies.
- No changes to `complyctl` or the plugin protobuf interface.
- `conftest --namespace` flag is stable and documented.

## Testing Strategy

- Unit tests per package (generate, scan, results, config, server).
- Integration-level test: Generate with mapping → Scan with namespace filtering → verify Gemara IDs in response.
- Fallback test: Generate without mapping → Scan with `--all-namespaces` → verify identical behavior to current implementation.
- All tests use dependency injection via existing interfaces (`CommandRunner`, `ToolChecker`, `DataLoader`).
