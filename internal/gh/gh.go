package gh

import (
	"fmt"
	"os"

	"github.com/VonC/godbg/exit"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

var GHex *exit.Exit
var client *github.Client

func init() {
	token := os.Getenv("GITHUB_AUTH_TOKEN")
	if token == "" {
		print("!!! No OAuth token. Limited API rate 60 per hour. !!!\n\n")
		client = github.NewClient(nil)
	} else {
		tc := oauth2.NewClient(oauth2.NoContext, oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		))
		client = github.NewClient(tc)
	}
}

type Commit struct {
	c *github.Commit
}

func (c *Commit) String() string {
	return fmt.Sprintf("commit '%s' from '%s', date '%s'",
		*c.c.SHA, *c.c.Author.Name, c.c.Author.Date.Format("02 Jan 2006"))
}

func MustGetCommit(sha1 string) *Commit {
	commit, _, err := client.Git.GetCommit("git", "git", sha1)
	if err != nil {
		fmt.Printf("Unable to get commit '%s': err '%v'\n", sha1, err)
		GHex.Exit(1)
	}
	return &Commit{c: commit}
}

func DisplayRateLimit() {
	rate, _, err := client.RateLimits()
	if err != nil {
		fmt.Printf("Error fetching rate limit: %#v\n\n", err)
	} else {
		const layout = "15:04pm (MST)"
		tc := rate.Core.Reset.Time
		tcs := fmt.Sprintf("%s", tc.Format(layout))
		ts := rate.Search.Reset.Time
		tss := fmt.Sprintf("%s", ts.Format(layout))
		fmt.Printf("\nAPI Rate Core Limit: %d/%d (reset at %s) - Search Limit: %d/%d (reset at %s)\n",
			rate.Core.Remaining, rate.Core.Limit, tcs,
			rate.Search.Remaining, rate.Search.Limit, tss)
	}
}
