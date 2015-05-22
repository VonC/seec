package main

import (
	"fmt"
	"os"

	"github.com/google/go-github/github"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run seec.go <sha1>")
		os.Exit(0)
	}
	sha1 := os.Args[1]
	client := github.NewClient(nil)
	commit, _, err := client.Git.GetCommit("git", "git", sha1)
	if err != nil {
		fmt.Printf("Unable to get commit '%s': err '%v'\n", sha1, err)
		os.Exit(1)
	}
	if len(commit.Parents) != 2 {
		fmt.Printf("Sha1 '%s' has '%d' parent(s) instead of 2\n", sha1, len(commit.Parents))
	}
	clogin := login(*commit.Author.Email, client)
	parent := commit.Parents[1]
	fmt.Printf("%+v\n", parent)
	pcommit, _, err := client.Git.GetCommit("git", "git", *parent.SHA)
	if err != nil {
		fmt.Printf("Unable to get parent commit '%s': err '%v'\n", parent.SHA, err)
		os.Exit(1)
	}
	plogin := login(*pcommit.Author.Email, client)

	fmt.Printf("See [commit %s](https://github.com/git/git/commit/%s) by [%s](https://github.com/%s (`%s`)), %s.  \n",
		(*pcommit.SHA)[:7], *pcommit.SHA,
		*pcommit.Author.Name, plogin, plogin, pcommit.Author.Date.Format("02 Jan 2006"))
	fmt.Printf("<sup>(Merged by [%s](https://github.com/%s -- `%s` --) in [commit %s](https://github.com/git/git/commit/%s), %s)</sup>",
		*commit.Author.Name, clogin, clogin,
		sha1[:7], sha1, commit.Committer.Date.Format("02 Jan 2006"))
}

func login(email string, client *github.Client) string {
	opts := &github.SearchOptions{Order: "desc"}
	res, _, err := client.Search.Users(email, opts)
	if err != nil {
		fmt.Printf("Unable to search user '%s': err '%v'", email, err)
		os.Exit(1)
	}
	if *res.Total == 0 {
		return ""
	}
	return *res.Users[0].Login
}
