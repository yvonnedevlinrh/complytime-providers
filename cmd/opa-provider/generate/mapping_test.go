// SPDX-License-Identifier: Apache-2.0

package generate

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/complytime/complyctl/pkg/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadMapping_Valid(t *testing.T) {
	dir := t.TempDir()
	data := `{
		"version": "1",
		"mappings": [
			{"requirement_id": "CIS-K8S-5.2.6", "namespace": "kubernetes.run_as_root"},
			{"requirement_id": "CIS-K8S-5.4.1", "namespace": "kubernetes.resource_limits"}
		]
	}`
	require.NoError(t, os.WriteFile(filepath.Join(dir, MappingFileName), []byte(data), 0600))

	mapping, err := LoadMapping(dir)
	require.NoError(t, err)
	assert.Equal(t, "1", mapping.Version)
	assert.Len(t, mapping.Mappings, 2)
	assert.Equal(t, "CIS-K8S-5.2.6", mapping.Mappings[0].RequirementID)
	assert.Equal(t, "kubernetes.run_as_root", mapping.Mappings[0].Namespace)
}

func TestLoadMapping_MissingFile(t *testing.T) {
	dir := t.TempDir()
	_, err := LoadMapping(dir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "reading mapping file")
}

func TestLoadMapping_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, MappingFileName), []byte(""), 0600))

	_, err := LoadMapping(dir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parsing mapping file")
}

func TestLoadMapping_MalformedJSON(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, MappingFileName), []byte("{invalid"), 0600))

	_, err := LoadMapping(dir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parsing mapping file")
}

func TestLoadMapping_DuplicateRequirementID(t *testing.T) {
	dir := t.TempDir()
	data := `{
		"version": "1",
		"mappings": [
			{"requirement_id": "CIS-1", "namespace": "ns1"},
			{"requirement_id": "CIS-1", "namespace": "ns2"}
		]
	}`
	require.NoError(t, os.WriteFile(filepath.Join(dir, MappingFileName), []byte(data), 0600))

	_, err := LoadMapping(dir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate requirement_id")
}

func TestLoadMapping_EmptyRequirementID(t *testing.T) {
	dir := t.TempDir()
	data := `{"version": "1", "mappings": [{"requirement_id": "", "namespace": "ns1"}]}`
	require.NoError(t, os.WriteFile(filepath.Join(dir, MappingFileName), []byte(data), 0600))

	_, err := LoadMapping(dir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty requirement_id")
}

func TestLoadMapping_EmptyNamespace(t *testing.T) {
	dir := t.TempDir()
	data := `{"version": "1", "mappings": [{"requirement_id": "CIS-1", "namespace": ""}]}`
	require.NoError(t, os.WriteFile(filepath.Join(dir, MappingFileName), []byte(data), 0600))

	_, err := LoadMapping(dir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty namespace")
}

func TestMatchRequirements_AllMatched(t *testing.T) {
	mapping := &MappingFile{
		Version: "1",
		Mappings: []MappingEntry{
			{RequirementID: "CIS-1", Namespace: "kubernetes.run_as_root"},
			{RequirementID: "CIS-2", Namespace: "kubernetes.resource_limits"},
		},
	}
	configs := []provider.AssessmentConfiguration{
		{RequirementID: "CIS-1"},
		{RequirementID: "CIS-2"},
	}

	namespaces, reverseMap, warnings := MatchRequirements(configs, mapping)
	assert.Len(t, warnings, 0)
	assert.Equal(t, []string{"kubernetes.resource_limits", "kubernetes.run_as_root"}, namespaces)
	assert.Equal(t, "CIS-1", reverseMap["kubernetes.run_as_root"])
	assert.Equal(t, "CIS-2", reverseMap["kubernetes.resource_limits"])
}

func TestMatchRequirements_PartialMatch(t *testing.T) {
	mapping := &MappingFile{
		Version: "1",
		Mappings: []MappingEntry{
			{RequirementID: "CIS-1", Namespace: "kubernetes.run_as_root"},
		},
	}
	configs := []provider.AssessmentConfiguration{
		{RequirementID: "CIS-1"},
		{RequirementID: "CIS-MISSING"},
	}

	namespaces, reverseMap, warnings := MatchRequirements(configs, mapping)
	assert.Len(t, warnings, 1)
	assert.Contains(t, warnings[0], "CIS-MISSING")
	assert.Equal(t, []string{"kubernetes.run_as_root"}, namespaces)
	assert.Equal(t, "CIS-1", reverseMap["kubernetes.run_as_root"])
}

func TestMatchRequirements_NoneMatched(t *testing.T) {
	mapping := &MappingFile{
		Version:  "1",
		Mappings: []MappingEntry{},
	}
	configs := []provider.AssessmentConfiguration{
		{RequirementID: "CIS-1"},
	}

	namespaces, _, warnings := MatchRequirements(configs, mapping)
	assert.Len(t, warnings, 1)
	assert.Empty(t, namespaces)
}

func TestMatchRequirements_DeduplicatesConfigs(t *testing.T) {
	mapping := &MappingFile{
		Version: "1",
		Mappings: []MappingEntry{
			{RequirementID: "CIS-1", Namespace: "ns1"},
		},
	}
	configs := []provider.AssessmentConfiguration{
		{RequirementID: "CIS-1"},
		{RequirementID: "CIS-1"},
	}

	namespaces, _, warnings := MatchRequirements(configs, mapping)
	assert.Len(t, warnings, 0)
	assert.Equal(t, []string{"ns1"}, namespaces)
}
