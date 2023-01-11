package gitrepo

import (
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

type ProjectGitData struct {
	Remotes []GitRemote `json:"remotes"`
}
type GitRemote struct {
	URL      string   `json:"url"`
	Branches []string `json:"branches"`
	Tags     []string `json:"tags"`
}
type GitData struct {
	Commit  string    `json:"commit"`
	Branch  string    `json:"branch"`
	Message string    `json:"message"`
	Author  string    `json:"author"`
	Time    time.Time `json:"time"`
	IsClean bool      `json:"clean"`
	Parents []string  `json:"parents"`
}

func GetProjectGitData(configDir string) (*ProjectGitData, error) {
	repo, err := git.PlainOpen(configDir)
	if err == git.ErrRepositoryNotExists {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	gitData := ProjectGitData{
		Remotes: []GitRemote{},
	}
	remotes, err := repo.Remotes()
	if err != nil {
		return nil, err
	}
	for _, remote := range remotes {
		remotePrefix := remote.Config().Name + "/"
		url := strings.TrimSuffix(remote.Config().URLs[0], ".git")
		branches := []string{}
		tags := []string{}

		refIter, _ := repo.References()
		refIter.ForEach(func(r *plumbing.Reference) error {
			if r.Name().IsRemote() {
				branch := r.Name().Short()
				if strings.HasPrefix(branch, remotePrefix) && !strings.HasSuffix(branch, "HEAD") {
					branches = append(branches, strings.TrimPrefix(branch, remotePrefix))
				}
			}
			if r.Name().IsTag() {
				tags = append(tags, r.Name().Short())
			}
			return nil
		})

		gitData.Remotes = append(gitData.Remotes, GitRemote{
			URL:      url,
			Branches: branches,
			Tags:     tags,
		})
	}

	return &gitData, nil
}

func GetGitData(configDir string) (*GitData, error) {
	repo, err := git.PlainOpen(configDir)
	if err == git.ErrRepositoryNotExists {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	ref, err := repo.Head()
	if err != nil {
		return nil, err
	}

	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return nil, err
	}

	parents := []string{}
	for _, p := range commit.ParentHashes {
		parents = append(parents, p.String())
	}

	w, err := repo.Worktree()
	if err != nil {
		return nil, err
	}
	status, err := w.Status()
	if err != nil {
		return nil, err
	}

	isClean := status.IsClean()

	return &GitData{
		IsClean: isClean,
		Commit:  commit.ID().String(),
		Message: strings.TrimSpace(commit.Message),
		Author:  commit.Author.String(),
		Time:    commit.Author.When,
		Parents: parents,
		Branch:  ref.Name().Short(),
	}, nil
}
