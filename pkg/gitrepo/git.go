package gitrepo

import (
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"golang.org/x/mod/semver"
)

type ProjectGitData struct {
	Remotes      []GitRemote   `json:"remotes"`
	Contributors []Contributor `json:"contributors"`
}
type Contributor struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Commits uint   `json:"commits"`
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
	Tags    []string  `json:"tags"`
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
		Remotes:      []GitRemote{},
		Contributors: []Contributor{},
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

		var mainBranch *plumbing.Reference

		refIter, _ := repo.References()
		refIter.ForEach(func(r *plumbing.Reference) error {
			if r.Name().IsRemote() {
				branch := r.Name().Short()
				if strings.HasPrefix(branch, remotePrefix) && !strings.HasSuffix(branch, "HEAD") {
					branches = append(branches, strings.TrimPrefix(branch, remotePrefix))
					if strings.HasSuffix(r.Name().Short(), "main") || strings.HasSuffix(r.Name().Short(), "master") {
						mainBranch = r
					}
				}
			}
			if r.Name().IsTag() {
				tags = append(tags, r.Name().Short())
			}
			return nil
		})

		contributors := map[string]Contributor{}
		if mainBranch != nil {
			cIter, err := repo.Log(&git.LogOptions{From: mainBranch.Hash()})
			if err != nil {
				return nil, err
			}
			cIter.ForEach(func(c *object.Commit) error {
				commiter := c.Committer.Email
				author := c.Author.Email

				contr, ok := contributors[commiter]
				if ok {
					contr.Commits += 1
					contributors[commiter] = contr
				} else {
					contributors[commiter] = Contributor{
						Name:    c.Committer.Name,
						Email:   c.Committer.Email,
						Commits: 1,
					}
				}

				if author != commiter {
					contr, ok := contributors[author]
					if ok {
						contr.Commits += 1
						contributors[author] = contr
					} else {
						contributors[author] = Contributor{
							Name:    c.Author.Name,
							Email:   c.Author.Email,
							Commits: 1,
						}
					}
				}
				return nil
			})
		}

		for _, contr := range contributors {
			gitData.Contributors = append(gitData.Contributors, contr)
		}

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

	tags := []string{}
	tagRefs, _ := repo.Tags()
	tagRefs.ForEach(func(t *plumbing.Reference) error {
		if t.Hash() == ref.Hash() {
			tagName := t.Name().Short()
			if semver.IsValid(tagName) {
				log.Warnf("Skipping an invalid tag %s that is not a valid semantic version, please check https://semver.org/", tagName)
				tags = append(tags, tagName)
			}
		}
		return nil
	})

	return &GitData{
		IsClean: isClean,
		Commit:  commit.ID().String(),
		Message: strings.TrimSpace(commit.Message),
		Author:  commit.Author.String(),
		Time:    commit.Author.When,
		Parents: parents,
		Branch:  ref.Name().Short(),
		Tags:    tags,
	}, nil
}
