package util

import (
	"os"
	"path/filepath"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

type GitUtil struct {
	URI  string
	Path string

	// Optional
	ReferenceName plumbing.ReferenceName
}

func NewGitUtil(uri string, path string, refName string) *GitUtil {
	cfg := &GitUtil{
		URI:           uri,
		Path:          path,
		ReferenceName: plumbing.HEAD,
	}
	if refName != "" {
		cfg.ReferenceName = plumbing.NewBranchReferenceName(refName)
	}
	return cfg
}

func (g *GitUtil) EnsureCloned() error {
	if ok, err := g.IsGitCloned(); err != nil {
		return err
	} else if !ok {
		_, err := git.PlainClone(g.Path, false, &git.CloneOptions{
			URL:           g.URI,
			ReferenceName: g.ReferenceName,
		})
		if err != nil && err != git.ErrRepositoryAlreadyExists {
			return err
		}
	}
	return nil
}

func (g *GitUtil) IsGitCloned() (bool, error) {
	f, err := os.Stat(filepath.Join(g.Path, ".git"))
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil && f.IsDir(), err
}

// EnsureUpdated will ensure the destination path exists and is up to date.
func (g *GitUtil) EnsureUpdated() error {
	if err := g.EnsureCloned(); err != nil {
		return err
	}
	return g.updateAndCleanUntracked()
}

func (g *GitUtil) updateAndCleanUntracked() error {
	repo, err := git.PlainOpen(g.Path)
	if err != nil {
		return err
	}
	err = repo.Fetch(&git.FetchOptions{})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return err
	}
	wt, err := repo.Worktree()
	if err != nil {
		return err
	}
	err = wt.Pull(&git.PullOptions{})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return err
	}
	err = wt.Reset(&git.ResetOptions{
		Commit: plumbing.ZeroHash,
		Mode:   git.HardReset,
	})
	if err != nil {
		return err
	}
	return wt.Clean(&git.CleanOptions{
		Dir: true,
	})
}