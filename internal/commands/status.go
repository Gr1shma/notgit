package commands

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Gr1shma/notgit/internal/repository"
	"github.com/spf13/cobra"
)

type FileStatus int

const (
	StatusUnmodified FileStatus = iota
	StatusModified
	StatusAdded
	StatusDeleted
)

type StatusEntry struct {
	Path          string
	IndexStatus   FileStatus
	WorkingStatus FileStatus
	OriginalPath  string // for renames -> TODO: implement it
}

type RepositoryStatus struct {
	Branch          string
	Entries         []StatusEntry
	UntrackedFiles  []string
	Repository      *repository.Repository
	HasChanges      bool
	StagedChanges   int
	UnstagedChanges int
}

type FileInfo struct {
	Path    string
	Hash    string
	Mode    os.FileMode
	ModTime int64
	Size    int64
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the working tree status",
	Long: `Show the status of files in the working directory and staging area.
This command shows which files have been modified, added, deleted, or are untracked.`,
	Args: cobra.NoArgs,
	RunE: statusCallback,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func statusCallback(cmd *cobra.Command, args []string) error {
	repo, err := repository.OpenRepository(".")
	if err != nil {
		return fmt.Errorf("not inside a notgit repository: %w", err)
	}

	status, err := getRepositoryStatus(repo)
	if err != nil {
		return fmt.Errorf("error while getting repository status: %w", err)
	}

	printStatus(cmd, status)
	return nil
}

func getRepositoryStatus(repo *repository.Repository) (*RepositoryStatus, error) {
	repoStatus := &RepositoryStatus{
		Repository: repo,
	}

	repoCurrentBranch, err := repo.GetCurrentBranch()
	if err != nil {
		return nil, fmt.Errorf("error: getting current branch: %v", err)
	}
	repoStatus.Branch = repoCurrentBranch

	latestCommitTree, err := getLatestCommitTree(repo)
	if err != nil {
		return nil, fmt.Errorf("error: getting latest commit tree: %v", err)
	}

	indexFiles, err := getIndexFiles(repo)
	if err != nil {
		return nil, fmt.Errorf("error: getting index files: %v", err)
	}

	workingFiles, err := getWorkingDirectoryFiles(repo)
	if err != nil {
		return nil, fmt.Errorf("error: getting working directory files: %v", err)
	}

	entries, untracked := buildCompleteStatusEntries(latestCommitTree, indexFiles, workingFiles)
	repoStatus.Entries = entries
	repoStatus.UntrackedFiles = untracked

	for _, entry := range entries {
		if entry.IndexStatus != StatusUnmodified {
			repoStatus.StagedChanges++
		}
		if entry.WorkingStatus != StatusUnmodified {
			repoStatus.UnstagedChanges++
		}
	}

	repoStatus.HasChanges = repoStatus.StagedChanges > 0 || repoStatus.UnstagedChanges > 0 || len(repoStatus.UntrackedFiles) > 0

	return repoStatus, nil
}

func getLatestCommitTree(repo *repository.Repository) (map[string]string, error) {
	headPath := filepath.Join(repo.NotgitDir, "HEAD")
	headContent, err := os.ReadFile(headPath)
	if err != nil {
		return nil, fmt.Errorf("error while reading the HEAD file")
	}

	ref := strings.TrimSpace(string(headContent))
	var commitHash string
	if strings.HasPrefix(ref, "ref: ") {
		refPath := filepath.Join(repo.NotgitDir, strings.TrimPrefix(ref, "ref: "))
		commitHashBytes, err := os.ReadFile(refPath)
		if err != nil {
			if os.IsNotExist(err) {
				return make(map[string]string), nil
			}
			return nil, fmt.Errorf("failed to read ref file: %v", err)
		}
		commitHash = strings.TrimSpace(string(commitHashBytes))
	} else {
		commitHash = ref
	}
	if commitHash == "" {
		return make(map[string]string), nil
	}

	commit, err := repo.RetrieveCommit(commitHash)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve commit: %v", err)
	}

	tree, err := repo.RetrieveTree(commit.TreeHash)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve tree: %v", err)
	}

	treeFiles := make(map[string]string)
	for _, entry := range tree.Entries {
		treeFiles[entry.Name] = entry.Hash
	}

	return treeFiles, nil
}

func getIndexFiles(repo *repository.Repository) (map[string]FileInfo, error) {
	index, err := repo.LoadIndex()
	if err != nil {
		return nil, fmt.Errorf("failed to load index: %w", err)
	}
	files := make(map[string]FileInfo)
	for path, entry := range index.Entries {
		files[path] = FileInfo{
			Path: entry.Path,
			Hash: entry.Hash,
		}
	}
	return files, nil
}

func getWorkingDirectoryFiles(repo *repository.Repository) (map[string]FileInfo, error) {
	files := make(map[string]FileInfo)
	err := filepath.Walk(repo.BaseDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(repo.BaseDir, path)
		if err != nil {
			return err
		}

		if strings.HasPrefix(relPath, ".notgit") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if info.IsDir() || relPath == "." {
			return nil
		}

		relPath = filepath.ToSlash(relPath)

		hash, err := computeFileHash(path)
		if err != nil {
			return nil
		}

		files[relPath] = FileInfo{
			Path:    relPath,
			Hash:    hash,
			Mode:    info.Mode(),
			ModTime: info.ModTime().Unix(),
			Size:    info.Size(),
		}

		return nil
	})

	return files, err
}

func computeFileHash(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	content := fmt.Sprintf("blob %d\x00%s", len(data), string(data))

	hash := sha1.Sum([]byte(content))
	return hex.EncodeToString(hash[:]), nil
}

func buildCompleteStatusEntries(headTree map[string]string, indexFiles map[string]FileInfo, workingFiles map[string]FileInfo) ([]StatusEntry, []string) {
	var entries []StatusEntry
	var untracked []string
	processedFiles := make(map[string]bool)

	allPaths := make(map[string]bool)
	for path := range headTree {
		allPaths[path] = true
	}
	for path := range indexFiles {
		allPaths[path] = true
	}
	for path := range workingFiles {
		allPaths[path] = true
	}

	for path := range allPaths {
		entry := StatusEntry{Path: path}

		headHash, inHead := headTree[path]
		indexFile, inIndex := indexFiles[path]
		workingFile, inWorking := workingFiles[path]

		// Determine index status (HEAD vs Index comparison)
		if !inHead && inIndex {
			// New file added to index
			entry.IndexStatus = StatusAdded
		} else if inHead && !inIndex {
			// File deleted from index
			entry.IndexStatus = StatusDeleted
		} else if inHead && inIndex {
			// File exists in both, check if modified
			if headHash != indexFile.Hash {
				entry.IndexStatus = StatusModified
			} else {
				entry.IndexStatus = StatusUnmodified
			}
		} else {
			// File doesn't exist in HEAD or Index
			entry.IndexStatus = StatusUnmodified
		}

		// Determine working status (Index vs Working Directory comparison)
		if !inIndex && inWorking {
			// File exists in working dir but not in index
			if !inHead {
				// File is completely untracked
				untracked = append(untracked, path)
				continue
			} else {
				// File was deleted from index but still exists in working dir
				entry.WorkingStatus = StatusModified
			}
		} else if inIndex && !inWorking {
			// File deleted from working directory
			entry.WorkingStatus = StatusDeleted
		} else if inIndex && inWorking {
			// File exists in both, check if modified
			if indexFile.Hash != workingFile.Hash {
				entry.WorkingStatus = StatusModified
			} else {
				entry.WorkingStatus = StatusUnmodified
			}
		} else {
			// File doesn't exist in index or working dir
			entry.WorkingStatus = StatusUnmodified
		}

		// Only add entry if there are changes
		if entry.IndexStatus != StatusUnmodified || entry.WorkingStatus != StatusUnmodified {
			entries = append(entries, entry)
		}

		processedFiles[path] = true
	}

	// Sort entries and untracked files
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Path < entries[j].Path
	})
	sort.Strings(untracked)

	return entries, untracked
}

func printStatus(cmd *cobra.Command, status *RepositoryStatus) {
	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "On branch %s\n", status.Branch)

	if !status.HasChanges {
		fmt.Fprintf(out, "nothing to commit, working tree clean")
		return
	}

	stagedEntries := getStagedEntries(status.Entries)
	if len(stagedEntries) > 0 {
		fmt.Fprintf(out, "\nChanges to be committed:")
		fmt.Fprintf(out, "  (use \"notgit reset HEAD <file>...\" to unstage)\n")
		for _, entry := range stagedEntries {
			fmt.Fprintf(out, "  %s: %s\n", getStatusString(entry.IndexStatus), entry.Path)
		}
	}

	unstagedEntries := getUnstagedEntries(status.Entries)
	if len(unstagedEntries) > 0 {
		fmt.Fprintf(out, "\nChanges not staged for commit:")
		fmt.Fprintf(out, "  (use \"notgit add <file>...\" to update what will be committed)")
		fmt.Fprintf(out, "  (use \"notgit checkout -- <file>...\" to discard changes in working directory)\n")
		for _, entry := range unstagedEntries {
			fmt.Fprintf(out, "  %s: %s\n", getStatusString(entry.WorkingStatus), entry.Path)
		}
	}

	if len(status.UntrackedFiles) > 0 {
		fmt.Fprintf(out, "\nUntracked files:")
		fmt.Fprintf(out, "  (use \"notgit add <file>...\" to include in what will be committed)\n")
		for _, file := range status.UntrackedFiles {
			fmt.Fprintf(out, "  %s\n", file)
		}
	}

	if len(stagedEntries) == 0 && (len(unstagedEntries) > 0 || len(status.UntrackedFiles) > 0) {
		fmt.Fprintf(out, "\nno changes added to commit (use \"notgit add\" and/or \"notgit commit -a\")")
	}
}

func getStagedEntries(entries []StatusEntry) []StatusEntry {
	var staged []StatusEntry
	for _, entry := range entries {
		if entry.IndexStatus != StatusUnmodified {
			staged = append(staged, entry)
		}
	}
	sort.Slice(staged, func(i, j int) bool {
		return staged[i].Path < staged[j].Path
	})
	return staged
}

func getUnstagedEntries(entries []StatusEntry) []StatusEntry {
	var unstaged []StatusEntry
	for _, entry := range entries {
		if entry.WorkingStatus != StatusUnmodified {
			unstaged = append(unstaged, entry)
		}
	}
	sort.Slice(unstaged, func(i, j int) bool {
		return unstaged[i].Path < unstaged[j].Path
	})
	return unstaged
}

func getStatusString(status FileStatus) string {
	switch status {
	case StatusModified:
		return "modified"
	case StatusAdded:
		return "new file"
	case StatusDeleted:
		return "deleted"
	default:
		return "unknown"
	}
}
