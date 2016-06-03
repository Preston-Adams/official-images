package main

import (
	"fmt"

	"github.com/codegangsta/cli"
)

func cmdBuild(c *cli.Context) error {
	repos, err := repos(c.Bool("all"), c.Args()...)
	if err != nil {
		return cli.NewMultiError(fmt.Errorf(`failed gathering repo list`), err)
	}

	repos, err = sortRepos(repos)
	if err != nil {
		return cli.NewMultiError(fmt.Errorf(`failed sorting repo list`, err))
	}

	uniq := c.Bool("uniq")
	namespace := c.String("namespace")
	pullMissing := c.Bool("pull-missing")

	for _, repo := range repos {
		r, err := fetch(repo)
		if err != nil {
			return cli.NewMultiError(fmt.Errorf(`failed fetching repo %q`, repo), err)
		}

		entries, err := r.SortedEntries()
		if err != nil {
			return cli.NewMultiError(fmt.Errorf(`failed sorting entries list for %q`, repo), err)
		}

		for _, entry := range entries {
			if r.SkipConstraints(entry) {
				continue
			}

			from, err := r.DockerFrom(&entry)
			if err != nil {
				return cli.NewMultiError(fmt.Errorf(`failed fetching/scraping FROM for %q (tags %q)`, r.RepoName, entry.TagsString()), err)
			}

			if pullMissing {
				_, err := dockerInspect("{{.Id}}", from)
				if err != nil {
					fmt.Printf("Pulling %s (%s)\n", from, r.RepoName)
					dockerPull(from)
				}
			}

			cacheHash, err := r.dockerCacheHash(&entry)
			if err != nil {
				return cli.NewMultiError(fmt.Errorf(`failed calculating "cache hash" for %q (tags %q)`, r.RepoName, entry.TagsString()), err)
			}

			cacheTag := "bashbrew/cache:" + cacheHash

			// check whether we've already built this artifact
			_, err = dockerInspect("{{.Id}}", cacheTag)
			if err != nil {
				fmt.Printf("Building %s (%s)\n", cacheTag, r.RepoName)

				commit, err := r.fetchGitRepo(&entry)
				if err != nil {
					return cli.NewMultiError(fmt.Errorf(`failed fetching git repo for %q (tags %q)`, r.RepoName, entry.TagsString()), err)
				}

				archive, err := gitArchive(commit, entry.Directory)
				if err != nil {
					return cli.NewMultiError(fmt.Errorf(`failed generating git archive for %q (tags %q)`, r.RepoName, entry.TagsString()), err)
				}
				defer archive.Close()

				err = dockerBuild(cacheTag, archive)
				if err != nil {
					return cli.NewMultiError(fmt.Errorf(`failed building %q (tags %q)`, r.RepoName, entry.TagsString()), err)
				}
			}

			for _, tag := range r.Tags(namespace, uniq, entry) {
				fmt.Printf("Tagging %s\n", tag)

				err := dockerTag(cacheTag, tag)
				if err != nil {
					return cli.NewMultiError(fmt.Errorf(`failed tagging %q as %q`, cacheTag, tag), err)
				}
			}
		}
	}

	return nil
}
