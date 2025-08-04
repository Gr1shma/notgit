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
	StatusRenamed
	StatusCopied
	StatusUnmerged
	StatusUntracked
	StatusIgnored
)

type StatusEntry struct {
	Path          string
	IndexStatus   FileStatus
	WorkingStatus FileStatus
	OriginalPath  string // for renames
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
	Run:  statusCallback,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func statusCallback(cmd *cobra.Command, args []string) {
	repo, err := repository.OpenRepository(".")
	if err != nil {
		fmt.Printf("Error: not inside a notgit repository: %v\n", err)
		return
	}

	status, err := getRepositoryStatus(repo)
	if err != nil {
		fmt.Printf("Error: getting repository status: %v\n", err)
		return
	}
	printStatus(status)
}

func getRepositoryStatus(repo *repository.Repository) (*RepositoryStatus, error) {
	repoStatus := &RepositoryStatus{
		Repository: repo,
	}

	repoCurrentBranch, err := getBranch(repo)
	if err != nil {
		fmt.Printf("Error: getting current branch: %v\n", err)
	} else {
		repoStatus.Branch = repoCurrentBranch
	}

	indexFiles, err := getIndexFiles(repo)
	if err != nil {
		return nil, fmt.Errorf("failed to read index: %w", err)
	}

	workingFiles, err := getWorkingDirectoryFiles(repo)
	if err != nil {
		return nil, fmt.Errorf("failed to scan working directory: %w", err)
	}

	repoStatus.Entries = buildStatusEntries(indexFiles, workingFiles)
	repoStatus.UntrackedFiles = getUntrackedFiles(indexFiles, workingFiles)
	for _, entry := range repoStatus.Entries {
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

func getBranch(repo *repository.Repository) (string, error) {
	headPath := filepath.Join(repo.NotgitDir, "HEAD")
	headData, err := os.ReadFile(headPath)
	if err != nil {
		return "", err
	}
	headContent := strings.TrimSpace(string(headData))

	if branch, ok := strings.CutPrefix(headContent, "ref: refs/heads/"); ok {
		return branch, nil
	}

	if len(headContent) == 40 {
		return headContent[:7], nil // show short SHA
	}

	return "unknown", nil
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

func buildStatusEntries(indexFiles map[string]FileInfo, workingFiles map[string]FileInfo) []StatusEntry {
	var entries []StatusEntry
	processedFiles := make(map[string]bool)

	for path, indexFile := range indexFiles {
		entry := StatusEntry{
			Path: path,
		}
		processedFiles[path] = true
		if workingFile, exists := workingFiles[path]; exists {
			if indexFile.Hash != workingFile.Hash {
				entry.WorkingStatus = StatusModified
			}
		} else {
			entry.WorkingStatus = StatusDeleted
		}

		entry.IndexStatus = StatusUnmodified

		entries = append(entries, entry)
	}

	for path, workingFile := range workingFiles {
		if processedFiles[path] {
			continue
		}
		if indexFile, exists := indexFiles[path]; exists {
			if indexFile.Hash != workingFile.Hash {
				entries = append(entries, StatusEntry{
					Path:          path,
					WorkingStatus: StatusModified,
				})
			}
		}
		processedFiles[path] = true
	}
	return entries
}

func getUntrackedFiles(indexFiles, workingFiles map[string]FileInfo) []string {
	var untracked []string

	for path := range workingFiles {
		if _, exists := indexFiles[path]; !exists {
			untracked = append(untracked, path)
		}
	}

	sort.Strings(untracked)
	return untracked
}

func printStatus(status *RepositoryStatus) {
	fmt.Printf("On branch %s\n", status.Branch)

	if !status.HasChanges {
		fmt.Println("nothing to commit, working tree clean")
		return
	}

	stagedEntries := getStagedEntries(status.Entries)
	if len(stagedEntries) > 0 {
		fmt.Println("\nChanges to be committed:")
		fmt.Println()
		for _, entry := range stagedEntries {
			fmt.Printf("  %s: %s\n", getStatusString(entry.IndexStatus), entry.Path)
		}
	}

	unstagedEntries := getUnstagedEntries(status.Entries)
	if len(unstagedEntries) > 0 {
		fmt.Println("\nChanges not staged for commit:")
		fmt.Println("  use \"notgit add <file>\" to update what will be committed")
		fmt.Println()
		for _, entry := range unstagedEntries {
			fmt.Printf("  %s: %s\n", getStatusString(entry.WorkingStatus), entry.Path)
		}
	}

	if len(status.UntrackedFiles) > 0 {
		fmt.Println("\nUntracked files:")
		fmt.Println("  use \"notgit add <file>\" to include in what will be committed")
		fmt.Println()
		for _, file := range status.UntrackedFiles {
			fmt.Printf("  %s\n", file)
		}
	}

	fmt.Println()
	if status.StagedChanges == 0 && status.UnstagedChanges == 0 {
		fmt.Println("no changes added to commit use \"notgit add\"")
	} else if status.StagedChanges == 0 {
		fmt.Println("no changes added to commit use \"notgit add\"")
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
	case StatusRenamed:
		return "renamed"
	case StatusCopied:
		return "copied"
	case StatusUntracked:
		return "untracked"
	default:
		return "unknown"
	}
}
