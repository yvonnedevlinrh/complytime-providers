// SPDX-License-Identifier: Apache-2.0

package generate

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/complytime/complyctl/pkg/provider"
)

const (
	// MappingFileName is the name of the mapping file inside an OCI policy bundle.
	MappingFileName = "complytime-mapping.json"
)

// MappingFile represents the complytime-mapping.json file shipped in an OCI bundle.
type MappingFile struct {
	Version  string         `json:"version"`
	Mappings []MappingEntry `json:"mappings"`
}

// MappingEntry maps a Gemara assessment plan RequirementID to a Rego namespace.
type MappingEntry struct {
	RequirementID string `json:"requirement_id"`
	Namespace     string `json:"namespace"`
}

// LoadMapping reads and validates a complytime-mapping.json file from the given
// bundle directory. Returns an error if the file is missing, malformed, or
// contains invalid entries.
func LoadMapping(bundleDir string) (*MappingFile, error) {
	path := filepath.Join(bundleDir, MappingFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading mapping file: %w", err)
	}

	var mapping MappingFile
	if err := json.Unmarshal(data, &mapping); err != nil {
		return nil, fmt.Errorf("parsing mapping file: %w", err)
	}

	if err := validateMapping(&mapping); err != nil {
		return nil, fmt.Errorf("validating mapping file: %w", err)
	}

	return &mapping, nil
}

// validateMapping checks that the mapping file has no empty or duplicate entries.
func validateMapping(m *MappingFile) error {
	seen := make(map[string]bool, len(m.Mappings))
	for i, entry := range m.Mappings {
		if entry.RequirementID == "" {
			return fmt.Errorf("mapping entry %d has empty requirement_id", i)
		}
		if entry.Namespace == "" {
			return fmt.Errorf("mapping entry %d has empty namespace", i)
		}
		if seen[entry.RequirementID] {
			return fmt.Errorf("duplicate requirement_id %q in mapping", entry.RequirementID)
		}
		seen[entry.RequirementID] = true
	}
	return nil
}

// MatchRequirements matches assessment plan RequirementIDs against mapping
// entries using exact string equality (like AMPEL's MatchPolicies). Returns
// the matched namespace list, a reverse mapping for result ID resolution,
// and any warnings for unmatched requirements.
func MatchRequirements(
	configs []provider.AssessmentConfiguration,
	mapping *MappingFile,
) (namespaces []string, reverseMap map[string]string, warnings []string) {
	// Build lookup map: requirement_id -> namespace
	lookup := make(map[string]string, len(mapping.Mappings))
	for _, entry := range mapping.Mappings {
		lookup[entry.RequirementID] = entry.Namespace
	}

	seen := make(map[string]bool)
	reverseMap = make(map[string]string)

	for _, cfg := range configs {
		reqID := cfg.RequirementID
		if seen[reqID] {
			continue
		}
		seen[reqID] = true

		ns, ok := lookup[reqID]
		if !ok {
			warnings = append(warnings,
				fmt.Sprintf("no mapping found for requirement %q", reqID))
			continue
		}
		namespaces = append(namespaces, ns)
		reverseMap[ns] = reqID
	}

	sort.Strings(namespaces)
	return namespaces, reverseMap, warnings
}
