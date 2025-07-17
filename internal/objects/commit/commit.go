package commit

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/Gr1shma/notgit/internal/objects"
)

type Commit struct {
	TreeHash   string
	ParentHash []string
	Author     Signature
	Committer  Signature
	Message    string
	Hash       string
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

func (c *Commit) GetHash() string {
	return c.Hash
}

func NewCommit(treeHash, message string, parentHash []string, author, committer Signature) *Commit {
	return &Commit{
		TreeHash:   treeHash,
		ParentHash: parentHash,
		Author:     author,
		Committer:  committer,
		Message:    message,
	}
}

func (c *Commit) Serialize() ([]byte, error) {
	var buffer bytes.Buffer

	buffer.WriteString(fmt.Sprintf("tree %s\n", c.TreeHash))

	for _, parent := range c.ParentHash {
		buffer.WriteString(fmt.Sprintf("parent %s\n", parent))
	}

	buffer.WriteString(fmt.Sprintf("author %s <%s> %d %s\n",
		c.Author.Name,
		c.Author.Email,
		c.Author.Time.Unix(),
		c.Author.Time.Format("-0700"),
	))

	buffer.WriteString(fmt.Sprintf("commiter %s <%s> %d %s\n",
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

	hash := sha1.Sum(full)
	c.Hash = hex.EncodeToString(hash[:])

	return full, nil
}
