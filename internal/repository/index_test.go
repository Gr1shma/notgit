package repository_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Gr1shma/notgit/internal/repository"
	"github.com/stretchr/testify/require"
)

func TestNewIndex(t *testing.T) {
	tempDir := t.TempDir()
	notgitDir := filepath.Join(tempDir, ".notgit")
	err := os.MkdirAll(notgitDir, 0o755)
	require.NoError(t, err)

	repo := &repository.Repository{
		BaseDir:   tempDir,
		NotgitDir: notgitDir,
	}

	index := repository.NewIndex()
	index.AddEntry("file.txt", "hash")
	index.AddEntry("subdir/file.txt", "anotherhash")

	err = repo.SaveIndex(index)
	require.NoError(t, err)

	loadedIndex, err := repo.LoadIndex()
	require.NoError(t, err)

	require.Len(t, loadedIndex.Entries, 2)
	require.Equal(t, "hash", loadedIndex.Entries["file.txt"].Hash)
	require.Equal(t, "anotherhash", loadedIndex.Entries["subdir/file.txt"].Hash)
}
