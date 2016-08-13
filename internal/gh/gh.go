package gh

import (
	"fmt"
	"os"

	"github.com/VonC/godbg/exit"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

var GHex *exit.Exit
var Client *github.Client

func init() {
	token := os.Getenv("GITHUB_AUTH_TOKEN")
	if token == "" {
		print("!!! No OAuth token. Limited API rate 60 per hour. !!!\n")
		print("Set GITHUB_AUTH_TOKEN environment variable to your GitHub PTA\n\n")
		Client = github.NewClient(nil)
	} else {
		tc := oauth2.NewClient(oauth2.NoContext, oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		))
		Client = github.NewClient(tc)
	}
}

type Commit struct {
	*github.Commit
	authorDate string
}

func NewCommit(ghc *github.Commit) *Commit {
	return &Commit{ghc, ""}
}

func (c *Commit) String() string {
	f := ""
	if c.Author != nil {
		f = fmt.Sprintf(" from '%s', date '%s'",
			*c.Author.Name, c.Author.Date.Format("02 Jan 2006"))
	}
	return fmt.Sprintf("commit '%s'%s",
		*c.SHA, f)
}

func (c *Commit) NbParents() int {
	return len(c.Parents)
}

func (c *Commit) AuthorDate() string {
	if c.authorDate != "" {
		return c.authorDate
	}
	if c.Message == nil {
		c.Commit = MustGetCommit(*c.SHA).Commit
	}
	c.authorDate = c.Author.Date.Format("02 Jan 2006")
	return c.authorDate
}

func (c *Commit) CommitterDate() string {
	return c.Committer.Date.Format("02 Jan 2006")
}

func (c *Commit) FirstParent() *Commit {
	return NewCommit(&c.Parents[0])
}
func (c *Commit) SecondParent() *Commit {
	return NewCommit(&c.Parents[1])
}

func (c *Commit) SameSHA1(c2 *Commit) bool {
	return *c.SHA == *c2.SHA
}

func (c *Commit) SameAuthor(c2 *Commit) bool {
	return *c.Author.Name == *c2.Author.Name
}

func (c *Commit) MessageC() string {
	if c.Message == nil {
		c.Commit = MustGetCommit(*c.SHA).Commit
	}
	return *c.Message
}

func (c *Commit) AuthorName() string {
	return *c.Author.Name
}

func MustGetCommit(sha1 string) *Commit {
	commit, _, err := Client.Git.GetCommit("git", "git", sha1)
	if err != nil {
		fmt.Printf("Unable to get commit '%s': err '%v'\n", sha1, err)
		GHex.Exit(1)
	}
	return NewCommit(commit)
}

func FirstSingleParentCommit(parent *Commit) *Commit {
	var pcommit *github.Commit
	var err error
	for pcommit == nil {
		pcommit, _, err = Client.Git.GetCommit("git", "git", *parent.SHA)
		if err != nil {
			fmt.Printf("Unable to get parent commit '%s': err '%v'\n", parent.SHA, err)
			GHex.Exit(1)
		}
		// fmt.Printf("pcommit '%+v', len %d\n", pcommit, len(pcommit.Parents))
		if len(pcommit.Parents) == 2 {
			parent = NewCommit(&pcommit.Parents[1])
			pcommit = nil
		} else {
			break
		}
	}
	return NewCommit(pcommit)
}

func DisplayRateLimit() {
	rate, _, err := Client.RateLimits()
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
