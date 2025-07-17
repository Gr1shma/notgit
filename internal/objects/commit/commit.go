package commit

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Gr1shma/notgit/internal/objects"
)

type Commit struct {
	TreeHash     string
	ParentHashes []string
	Author       Signature
	Committer    Signature
	Message      string
	Hash         string
}

type Signature struct {
	Name  string
	Email string
	Time  time.Time
}

var _ objects.Object = (*Commit)(nil)

func (c *Commit) Type() string {
	return "Commit"
}

func NewCommit(treeHash, message string, parentHashes []string, author, committer Signature) *Commit {
	return &Commit{
		TreeHash:     treeHash,
		ParentHashes: parentHashes,
		Author:       author,
		Committer:    committer,
		Message:      message,
	}
}

func (c *Commit) Serialize() ([]byte, error) {
	var buffer bytes.Buffer

	buffer.WriteString(fmt.Sprintf("tree %s\n", c.TreeHash))

	for _, parent := range c.ParentHashes {
		buffer.WriteString(fmt.Sprintf("parent %s\n", parent))
	}

	buffer.WriteString(fmt.Sprintf("author %s <%s> %d %s\n",
		c.Author.Name,
		c.Author.Email,
		c.Author.Time.Unix(),
		c.Author.Time.Format("-0700"), // Git-style timezone offset
	))

	buffer.WriteString(fmt.Sprintf("committer %s <%s> %d %s\n",
		c.Committer.Name,
		c.Committer.Email,
		c.Committer.Time.Unix(),
		c.Committer.Time.Format("-0700"),
	))

	buffer.WriteString("\n")
	buffer.WriteString(c.Message)

	content := buffer.Bytes()
	header := fmt.Sprintf("commit %d\x00", len(content))
	full := append([]byte(header), content...)

	return full, nil
}

func (c *Commit) ComputeHash() error {
	serialized, err := c.Serialize()
	if err != nil {
		return err
	}
	hash := sha1.Sum(serialized)
	c.Hash = hex.EncodeToString(hash[:])
	return nil
}

func (c *Commit) GetHash() string {
	return c.Hash
}

func DeserializeCommit(data []byte) (*Commit, error) {
	nullByteIndex := bytes.IndexByte(data, 0)
	if nullByteIndex == -1 {
		return nil, fmt.Errorf("invalid commit format: missing null byte separator")
	}

	content := data[nullByteIndex+1:]
	lines := bytes.Split(content, []byte("\n"))

	commit := &Commit{}

	for i, lineBytes := range lines {
		line := string(lineBytes)

		if line == "" {
			commit.Message = string(bytes.Join(lines[i+1:], []byte("\n")))
			break
		}

		switch {
		case strings.HasPrefix(line, "tree "):
			commit.TreeHash = strings.TrimPrefix(line, "tree ")

		case strings.HasPrefix(line, "parent "):
			parentHash := strings.TrimPrefix(line, "parent ")
			commit.ParentHashes = append(commit.ParentHashes, parentHash)

		case strings.HasPrefix(line, "author "):
			sig, err := parseSignature(line[len("author "):])
			if err != nil {
				return nil, fmt.Errorf("invalid author line: %w", err)
			}
			commit.Author = sig

		case strings.HasPrefix(line, "committer "):
			sig, err := parseSignature(line[len("committer "):])
			if err != nil {
				return nil, fmt.Errorf("invalid committer line: %w", err)
			}
			commit.Committer = sig
		}
	}

	if err := commit.ComputeHash(); err != nil {
		return nil, err
	}

	return commit, nil
}

func parseSignature(input string) (Signature, error) {
	var sig Signature

	nameEmailParts := strings.SplitN(input, " <", 2)
	if len(nameEmailParts) != 2 {
		return sig, fmt.Errorf("invalid signature format: missing '<'")
	}
	sig.Name = nameEmailParts[0]

	emailAndTime := strings.SplitN(nameEmailParts[1], "> ", 2)
	if len(emailAndTime) != 2 {
		return sig, fmt.Errorf("invalid signature format: missing '> '")
	}
	sig.Email = emailAndTime[0]

	timeParts := strings.SplitN(emailAndTime[1], " ", 2)
	if len(timeParts) != 2 {
		return sig, fmt.Errorf("invalid time format")
	}
	unixTimestamp, err := strconv.ParseInt(timeParts[0], 10, 64)
	if err != nil {
		return sig, fmt.Errorf("invalid timestamp: %w", err)
	}

	sig.Time = time.Unix(unixTimestamp, 0)
	return sig, nil
}
