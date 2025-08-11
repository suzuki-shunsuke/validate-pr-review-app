package validation

import (
	"context"
	"io"

	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/enforce-pr-review-app/pkg/github"
)

type Controller struct {
	fs     afero.Fs
	stdout io.Writer
	stderr io.Writer
	gh     GitHub
}

type GitHub interface {
	GetPR(ctx context.Context, owner, repo string, number int) (*github.PullRequest, error)
}

func (c *Controller) Init(fs afero.Fs, gh GitHub, stdout, stderr io.Writer) {
	c.fs = fs
	c.gh = gh
	c.stdout = stdout
	c.stderr = stderr
}

type Input struct {
	RepoOwner             string
	RepoName              string
	PR                    int
	TrustedApps           map[string]struct{}
	UntrustedMachineUsers map[string]struct{}
}
