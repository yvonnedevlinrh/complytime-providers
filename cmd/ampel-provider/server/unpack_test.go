// SPDX-License-Identifier: Apache-2.0

package server

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// createTarGz builds a gzip-compressed tar archive in memory from the
// given entries. Each entry is a tarEntry describing a file, directory,
// or symlink to include.
type tarEntry struct {
	Name     string
	Body     []byte
	Typeflag byte
	Linkname string
}

func createTarGz(t *testing.T, entries []tarEntry) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)

	for _, e := range entries {
		hdr := &tar.Header{
			Name:     e.Name,
			Size:     int64(len(e.Body)),
			Typeflag: e.Typeflag,
			Mode:     0644,
			Linkname: e.Linkname,
		}
		if e.Typeflag == tar.TypeDir {
			hdr.Size = 0
		}
		require.NoError(t, tw.WriteHeader(hdr))
		if len(e.Body) > 0 {
			_, err := tw.Write(e.Body)
			require.NoError(t, err)
		}
	}

	require.NoError(t, tw.Close())
	require.NoError(t, gz.Close())
	return buf.Bytes()
}

func TestResolveComplypackPath_DirectoryPassthrough(t *testing.T) {
	dir := t.TempDir()
	policyDir := filepath.Join(dir, "policies")
	require.NoError(t, os.MkdirAll(policyDir, 0750))

	resolved, err := resolveComplypackPath(policyDir)
	require.NoError(t, err)
	require.Equal(t, policyDir, resolved)
}

func TestResolveComplypackPath_TarGzExtraction(t *testing.T) {
	dir := t.TempDir()

	archive := createTarGz(t, []tarEntry{
		{Name: "policy.json", Body: []byte(`{"id":"BP-1.01"}`), Typeflag: tar.TypeReg},
	})
	archivePath := filepath.Join(dir, "content.tar.gz")
	require.NoError(t, os.WriteFile(archivePath, archive, 0600))

	resolved, err := resolveComplypackPath(archivePath)
	require.NoError(t, err)

	expectedDir := filepath.Join(dir, "content")
	require.Equal(t, expectedDir, resolved)

	// Verify the extracted file exists with correct content.
	data, err := os.ReadFile(filepath.Join(expectedDir, "policy.json"))
	require.NoError(t, err)
	require.Contains(t, string(data), "BP-1.01")
}

func TestResolveComplypackPath_IdempotentSkip(t *testing.T) {
	dir := t.TempDir()

	archive := createTarGz(t, []tarEntry{
		{Name: "policy.json", Body: []byte(`{"id":"BP-1.01"}`), Typeflag: tar.TypeReg},
	})
	archivePath := filepath.Join(dir, "content.tar.gz")
	require.NoError(t, os.WriteFile(archivePath, archive, 0600))

	// Pre-create the content directory with a marker file.
	contentDir := filepath.Join(dir, "content")
	require.NoError(t, os.MkdirAll(contentDir, 0750))
	markerPath := filepath.Join(contentDir, "marker.txt")
	require.NoError(t, os.WriteFile(markerPath, []byte("existing"), 0600))

	resolved, err := resolveComplypackPath(archivePath)
	require.NoError(t, err)
	require.Equal(t, contentDir, resolved)

	// The marker file should still exist (no re-extraction).
	data, err := os.ReadFile(markerPath)
	require.NoError(t, err)
	require.Equal(t, "existing", string(data))
}

func TestExtractTarGz_PathTraversalRejected(t *testing.T) {
	dir := t.TempDir()

	archive := createTarGz(t, []tarEntry{
		{Name: "../escape.txt", Body: []byte("malicious"), Typeflag: tar.TypeReg},
	})
	archivePath := filepath.Join(dir, "bad.tar.gz")
	require.NoError(t, os.WriteFile(archivePath, archive, 0600))

	dst := filepath.Join(dir, "out")
	err := extractTarGz(archivePath, dst)
	require.Error(t, err)
	require.Contains(t, err.Error(), "path traversal")

	// The malicious file must not exist outside the destination.
	_, statErr := os.Stat(filepath.Join(dir, "escape.txt"))
	require.True(t, os.IsNotExist(statErr))
}

func TestExtractTarGz_SymlinkRejected(t *testing.T) {
	dir := t.TempDir()

	archive := createTarGz(t, []tarEntry{
		{Name: "link.txt", Typeflag: tar.TypeSymlink, Linkname: "/etc/passwd"},
	})
	archivePath := filepath.Join(dir, "symlink.tar.gz")
	require.NoError(t, os.WriteFile(archivePath, archive, 0600))

	dst := filepath.Join(dir, "out")
	err := extractTarGz(archivePath, dst)
	require.Error(t, err)
	require.Contains(t, err.Error(), "symlinks and hard links")
}

func TestExtractTarGz_OversizedFileRejected(t *testing.T) {
	dir := t.TempDir()

	// Create a tar with a header claiming a file larger than the limit.
	// We don't actually write 100 MB; the header size triggers the check.
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)

	bigSize := maxExtractedFileSize + 1
	hdr := &tar.Header{
		Name:     "huge.bin",
		Size:     int64(bigSize),
		Typeflag: tar.TypeReg,
		Mode:     0644,
	}
	require.NoError(t, tw.WriteHeader(hdr))

	// Write just enough data to exceed the limit. We write
	// maxExtractedFileSize + 1 bytes of zeros.
	chunk := make([]byte, 1<<20) // 1 MB chunks
	written := int64(0)
	for written < int64(bigSize) {
		n := int64(len(chunk))
		if written+n > int64(bigSize) {
			n = int64(bigSize) - written
		}
		_, err := tw.Write(chunk[:n])
		require.NoError(t, err)
		written += n
	}

	require.NoError(t, tw.Close())
	require.NoError(t, gz.Close())

	archivePath := filepath.Join(dir, "huge.tar.gz")
	require.NoError(t, os.WriteFile(archivePath, buf.Bytes(), 0600))

	dst := filepath.Join(dir, "out")
	err := extractTarGz(archivePath, dst)
	require.Error(t, err)
	require.Contains(t, err.Error(), "exceeds maximum size")
}

func TestResolveComplypackPath_NonExistentPath(t *testing.T) {
	_, err := resolveComplypackPath("/nonexistent/path/content.tar.gz")
	require.Error(t, err)
	require.Contains(t, err.Error(), "stat")
}

func TestExtractTarGz_CorruptArchive(t *testing.T) {
	dir := t.TempDir()

	// Write garbage data as if it were a tar.gz file.
	archivePath := filepath.Join(dir, "corrupt.tar.gz")
	require.NoError(t, os.WriteFile(archivePath, []byte("not a gzip file"), 0600))

	dst := filepath.Join(dir, "out")
	err := extractTarGz(archivePath, dst)
	require.Error(t, err)
	require.Contains(t, err.Error(), "gzip reader")
}

func TestExtractTarGz_HardLinkRejected(t *testing.T) {
	dir := t.TempDir()

	archive := createTarGz(t, []tarEntry{
		{Name: "hardlink.txt", Typeflag: tar.TypeLink, Linkname: "target.txt"},
	})
	archivePath := filepath.Join(dir, "hardlink.tar.gz")
	require.NoError(t, os.WriteFile(archivePath, archive, 0600))

	dst := filepath.Join(dir, "out")
	err := extractTarGz(archivePath, dst)
	require.Error(t, err)
	require.Contains(t, err.Error(), "symlinks and hard links")
}

func TestExtractTarGz_AbsolutePathRejected(t *testing.T) {
	dir := t.TempDir()

	archive := createTarGz(t, []tarEntry{
		{Name: "/etc/passwd", Body: []byte("root"), Typeflag: tar.TypeReg},
	})
	archivePath := filepath.Join(dir, "abs.tar.gz")
	require.NoError(t, os.WriteFile(archivePath, archive, 0600))

	dst := filepath.Join(dir, "out")
	err := extractTarGz(archivePath, dst)
	require.Error(t, err)
	require.Contains(t, err.Error(), "path traversal")
}

func TestExtractTarGz_DotDirectoryEntry(t *testing.T) {
	dir := t.TempDir()

	archive := createTarGz(t, []tarEntry{
		{Name: "./", Typeflag: tar.TypeDir},
		{Name: "policy.json", Body: []byte(`{"id":"BP-1.01"}`), Typeflag: tar.TypeReg},
	})
	archivePath := filepath.Join(dir, "dot.tar.gz")
	require.NoError(t, os.WriteFile(archivePath, archive, 0600))

	dst := filepath.Join(dir, "out")
	require.NoError(t, extractTarGz(archivePath, dst))

	data, err := os.ReadFile(filepath.Join(dst, "policy.json"))
	require.NoError(t, err)
	require.Contains(t, string(data), "BP-1.01")
}

func TestExtractTarGz_FilePermissions(t *testing.T) {
	dir := t.TempDir()

	archive := createTarGz(t, []tarEntry{
		{Name: "subdir/", Typeflag: tar.TypeDir},
		{Name: "subdir/policy.json", Body: []byte(`{}`), Typeflag: tar.TypeReg},
	})
	archivePath := filepath.Join(dir, "perms.tar.gz")
	require.NoError(t, os.WriteFile(archivePath, archive, 0600))

	dst := filepath.Join(dir, "out")
	require.NoError(t, extractTarGz(archivePath, dst))

	fi, err := os.Stat(filepath.Join(dst, "subdir"))
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0750), fi.Mode().Perm())

	fi, err = os.Stat(filepath.Join(dst, "subdir", "policy.json"))
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0600), fi.Mode().Perm())
}
