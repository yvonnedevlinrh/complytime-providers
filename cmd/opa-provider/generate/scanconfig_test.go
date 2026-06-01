// SPDX-License-Identifier: Apache-2.0

package generate

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteAndReadScanConfig_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	ids := []string{"kubernetes.run_as_root", "kubernetes.resource_limits"}
	reverseMap := map[string]string{
		"kubernetes.run_as_root":     "CIS-K8S-5.2.6",
		"kubernetes.resource_limits": "CIS-K8S-5.4.1",
	}

	err := WriteScanConfig(dir, ids, reverseMap, "/path/to/bundle")
	require.NoError(t, err)

	cfg, err := ReadScanConfig(dir)
	require.NoError(t, err)
	assert.Equal(t, ids, cfg.IDs)
	assert.Equal(t, reverseMap, cfg.ReverseMapping)
	assert.Equal(t, "/path/to/bundle", cfg.BundleDir)
	assert.NotEmpty(t, cfg.GeneratedAt)
}

func TestWriteAndReadScanConfig_NullIDs(t *testing.T) {
	dir := t.TempDir()

	err := WriteScanConfig(dir, nil, nil, "/path/to/bundle")
	require.NoError(t, err)

	cfg, err := ReadScanConfig(dir)
	require.NoError(t, err)
	assert.Nil(t, cfg.IDs)
	assert.Nil(t, cfg.ReverseMapping)
	assert.Equal(t, "/path/to/bundle", cfg.BundleDir)
}

func TestReadScanConfig_MissingFile(t *testing.T) {
	dir := t.TempDir()
	_, err := ReadScanConfig(dir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "reading scan config")
}

func TestReadScanConfig_MalformedJSON(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, ScanConfigFileName), []byte("{invalid"), 0600,
	))

	_, err := ReadScanConfig(dir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parsing scan config")
}

func TestWriteScanConfig_CreatesDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "dir")
	err := WriteScanConfig(dir, nil, nil, "")
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(dir, ScanConfigFileName))
	require.NoError(t, err)
}
