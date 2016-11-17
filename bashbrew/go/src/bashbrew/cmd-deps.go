package main

import (
	"fmt"

	"github.com/codegangsta/cli"
	"pault.ag/go/topsort"
)

func cmdOffspring(c *cli.Context) error {
	return cmdFamily(false, c)
}

func cmdParents(c *cli.Context) error {
	return cmdFamily(true, c)
}

func cmdFamily(parents bool, c *cli.Context) error {
	depsRepos, err := repos(c.Bool("all"), c.Args()...)
	if err != nil {
		return cli.NewMultiError(fmt.Errorf(`failed gathering repo list`), err)
	}

	uniq := c.Bool("uniq")
	applyConstraints := c.Bool("apply-constraints")

	allRepos, err := repos(true)
	if err != nil {
		return cli.NewMultiError(fmt.Errorf(`failed gathering ALL repos list`), err)
	}

	// create network (all repos)
	network := topsort.NewNetwork()

	// add nodes
	for _, repo := range allRepos {
		r, err := fetch(repo)
		if err != nil {
			return cli.NewMultiError(fmt.Errorf(`failed fetching repo %q`, repo), err)
		}

		for _, entry := range r.Entries() {
			if applyConstraints && r.SkipConstraints(entry) {
				continue
			}

			for _, tag := range r.Tags("", false, entry) {
				network.AddNode(tag, entry)
			}
		}
	}

	// add edges
	for _, repo := range allRepos {
		r, err := fetch(repo)
		if err != nil {
			return cli.NewMultiError(fmt.Errorf(`failed fetching repo %q`, repo), err)
		}
		for _, entry := range r.Entries() {
			if applyConstraints && r.SkipConstraints(entry) {
				continue
			}

			from, err := r.DockerFrom(&entry)
			if err != nil {
				return cli.NewMultiError(fmt.Errorf(`failed fetching/scraping FROM for %q (tags %q)`, r.RepoName, entry.TagsString()), err)
			}
			for _, tag := range r.Tags("", false, entry) {
				network.AddEdge(from, tag)
			}
		}
	}

	// now the real work
	seen := map[*topsort.Node]bool{}
	for _, repo := range depsRepos {
		r, err := fetch(repo)
		if err != nil {
			return cli.NewMultiError(fmt.Errorf(`failed fetching repo %q`, repo), err)
		}

		for _, entry := range r.Entries() {
			if applyConstraints && r.SkipConstraints(entry) {
				continue
			}

			for _, tag := range r.Tags("", uniq, entry) {
				nodes := []*topsort.Node{}
				if parents {
					nodes = append(nodes, network.Get(tag).InboundEdges...)
				} else {
					nodes = append(nodes, network.Get(tag).OutboundEdges...)
				}
				for len(nodes) > 0 {
					node := nodes[0]
					nodes = nodes[1:]
					if seen[node] {
						continue
					}
					seen[node] = true
					fmt.Printf("%s\n", node.Name)
					if parents {
						nodes = append(nodes, node.InboundEdges...)
					} else {
						nodes = append(nodes, node.OutboundEdges...)
					}
				}
			}
		}
	}

	return nil
}
