package git

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/protocol/packp/sideband"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"os"
	"path/filepath"
)

func Clone(dir, url, branch string, Depth int, progress sideband.Progress) (string, error) {
	path := filepath.Join(os.Getenv("HOME"), ".ssh", "id_rsa")
	if os.Getenv("OS") == "Windows_NT" {
		path = filepath.Join(os.Getenv("HOMEDRIVE"), os.Getenv("HOMEPATH"), ".ssh", "id_rsa")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	var key *ssh.PublicKeys
	key, _ = ssh.NewPublicKeys("git", data, "")
	repository, err := git.PlainClone(dir, false, &git.CloneOptions{
		URL:           url,
		Auth:          key,
		ReferenceName: plumbing.NewBranchReferenceName(branch),
		Depth:         Depth,
		Progress:      progress,
	})
	if err != nil {
		return "", err
	}
	head, err := repository.Head()
	if err != nil {
		return "", err
	}
	commit, err := repository.CommitObject(head.Hash())
	if err != nil {
		return "", err
	}
	return commit.Hash.String(), nil
}
