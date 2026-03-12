package config

import (
	"fmt"
	"path"
)

type Insecure struct {
	AllowUnsignedCommits  bool     `json:"allow_unsigned_commits,omitempty" yaml:"allow_unsigned_commits"`
	UnsignedCommitAuthors []string `json:"unsigned_commit_authors,omitempty" yaml:"unsigned_commit_authors"`
}

func (ins *Insecure) Validate() error {
	if ins == nil {
		return nil
	}
	for _, pattern := range ins.UnsignedCommitAuthors {
		if _, err := path.Match(pattern, "foo"); err != nil {
			return fmt.Errorf("invalid unsigned_commit_authors pattern %q: %w", pattern, err)
		}
	}
	return nil
}
