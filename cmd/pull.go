package cmd

import (
	color "github.com/fatih/color"
	cobra "github.com/spf13/cobra"
	git "gopkg.in/src-d/go-git.v4"
)

func init() {
	var pullCmd = &cobra.Command{
		Use:   "pull",
		Short: "Pull all repositories",
		Run: func(cmd *cobra.Command, args []string) {
			repoLoop(runPull, "Pulling")
		},
	}

	rootCmd.AddCommand(pullCmd)
}

func updateRepoConfig(conf Configuration, repository *git.Repository) {
	repoConf, err := repository.Config()
	fatalIfError(err)

	section := repoConf.Raw.Section("user")
	section.SetOption("name", conf.Fullname)
	section.SetOption("email", conf.Email)
	err = repoConf.Validate()
	fatalIfError(err)

	err = repository.Storer.SetConfig(repoConf)
	fatalIfError(err)
}

func runPull(conf Configuration, repo Repo, status *StatusList) {
	var repository *git.Repository
	var err error

	if pathExists(repo.Dir) {
		repository, err = git.PlainOpen(repo.Dir)
		// If we get ErrRepositoryNotExists here, it means the repo is broken
		if err == git.ErrRepositoryNotExists {
			status.append(repo.Dir, color.RedString("Broken"))
			return
		}
		fatalIfError(err)

		workTree, err := repository.Worktree()
		fatalIfError(err)

		err = workTree.Pull(&git.PullOptions{RemoteName: git.DefaultRemoteName})

		if err == git.ErrNonFastForwardUpdate {
			status.append(repo.Dir, color.RedString("Non-fast-forward update"))
			return
		}

		if err == git.NoErrAlreadyUpToDate {
			// Ignore NoErrAlreadyUpToDate
			err = nil
		}

		fatalIfError(err)
	} else {
		repository, err = git.PlainClone(repo.Dir, false, &git.CloneOptions{URL: repo.URL})
		fatalIfError(err)
	}

	updateRepoConfig(conf, repository)

	status.append(repo.Dir, color.GreenString("OK"))
}
