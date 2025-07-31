package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Gr1shma/notgit/internal/objects/commit"
	"github.com/Gr1shma/notgit/internal/objects/tree"
	"github.com/Gr1shma/notgit/internal/repository"
	"github.com/Gr1shma/notgit/internal/utils"
	"github.com/spf13/cobra"
)

var commitMessageString string

var commitCmd = &cobra.Command{
	Use:   "commit -m <message>",
	Short: "Record changes to the repository",
	Long: `Stores the current contents of the index in a new commit object.
A commit message is required to describe the changes being recorded.`,
	Run: commitCallback,
}

func init() {
	commitCmd.Flags().StringVarP(&commitMessageString, "message", "m", "", "Commit message (required)")
	commitCmd.MarkFlagRequired("message")
	rootCmd.AddCommand(commitCmd)
}

func commitCallback(cmd *cobra.Command, args []string) {
	repoRoot, err := utils.FindRepoRoot(".")
	if err != nil {
		fmt.Fprintln(os.Stderr, "fatal: not a notgit repository (or any of the parent directories): .notgit")
		os.Exit(1)
	}

	authorName, authorEmail, err := getUserIdentity()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, "\nPlease configure your user identity using:")
		fmt.Fprintln(os.Stderr, "  notgit config --global user.name \"Your Name\"")
		fmt.Fprintln(os.Stderr, "  notgit config --global user.email \"you@example.com\"")
		os.Exit(1)
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("error while getting current working directory: %v", err)
		return
	}

	repo, err := repository.OpenRepository(cwd)
	if err != nil {
		fmt.Printf("error while getting notgit repository: %v", err)
		return
	}

	idx, err := repo.LoadIndex()
	if err != nil {
		fmt.Printf("error while getting index from the repository: %v", err)
		return
	}

	if len(idx.Entries) == 0 {
		fmt.Println("Nothing to commit, working tree clean")
		return
	}

	commitTree := tree.NewTree()
	for path, entry := range idx.Entries {
		commitTree.AddEntry(path, entry.Hash, tree.EntryTypeBlob)
	}
	if err := commitTree.ComputeHash(); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: failed to compute tree hash: %v\n", err)
		os.Exit(1)
	}
	treeSHA := commitTree.GetHash()

	_, err = repo.StoreObject(commitTree)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fatal: failed to write tree object: %v\n", err)
		os.Exit(1)
	}

	parentSHA, _ := getParentCommit(repoRoot)

	now := time.Now()
	authorSig := commit.Signature{
		Name:  authorName,
		Email: authorEmail,
		Time:  now,
	}

	parentHashes := []string{}
	if parentSHA != "" {
		parentHashes = append(parentHashes, parentSHA)
	}

	commitObj := commit.NewCommit(treeSHA, commitMessageString, parentHashes, authorSig, authorSig)

	if err := commitObj.ComputeHash(); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: failed to compute commit hash: %v\n", err)
		os.Exit(1)
	}
	commitSHA := commitObj.GetHash()

	_, err = repo.StoreObject(commitObj)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fatal: failed to write commit object: %v\n", err)
		os.Exit(1)
	}

	if err := updateHead(repoRoot, commitSHA); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: failed to update HEAD: %v\n", err)
		os.Exit(1)
	}

	if parentSHA == "" {
		fmt.Printf("[root-commit %s] %s\n", commitSHA[:7], commitMessageString)
	} else {
		headRef, _ := readHeadRef(repoRoot)
		branchName := strings.TrimPrefix(headRef, "refs/heads/")
		fmt.Printf("[%s %s] %s\n", branchName, commitSHA[:7], commitMessageString)
	}
}

func getUserIdentity() (name, email string, err error) {
	localCfg, _, localErr := utils.LoadConfig(false)
	if localErr == nil {
		name, _ = utils.GetConfigKeyValue(localCfg, "user.name")
		email, _ = utils.GetConfigKeyValue(localCfg, "user.email")
	}

	if name == "" || email == "" {
		globalCfg, _, globalErr := utils.LoadConfig(true)
		if globalErr == nil {
			if name == "" {
				name, _ = utils.GetConfigKeyValue(globalCfg, "user.name")
			}
			if email == "" {
				email, _ = utils.GetConfigKeyValue(globalCfg, "user.email")
			}
		}
	}

	if name == "" || email == "" {
		return "", "", fmt.Errorf("fatal: user identity not set")
	}
	return name, email, nil
}

func readHeadRef(repoRoot string) (string, error) {
	headPath := filepath.Join(repoRoot, ".notgit", "HEAD")
	content, err := os.ReadFile(headPath)
	if err != nil {
		return "", err
	}
	ref := strings.TrimSpace(string(content))
	if !strings.HasPrefix(ref, "ref: ") {
		return "", fmt.Errorf("invalid HEAD content: %s", ref)
	}
	return strings.TrimPrefix(ref, "ref: "), nil
}

func getParentCommit(repoRoot string) (string, error) {
	refPath, err := readHeadRef(repoRoot)
	if err != nil {
		return "", err
	}

	parentSHABytes, err := os.ReadFile(filepath.Join(repoRoot, ".notgit", refPath))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(parentSHABytes)), nil
}

func updateHead(repoRoot, commitSHA string) error {
	refPath, err := readHeadRef(repoRoot)
	if err != nil {
		return err
	}
	fullRefPath := filepath.Join(repoRoot, ".notgit", refPath)
	if err := os.MkdirAll(filepath.Dir(fullRefPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(fullRefPath, []byte(commitSHA+"\n"), 0o644)
}
