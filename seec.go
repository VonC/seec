package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/VonC/godbg/exit"
	"github.com/atotto/clipboard"
	"github.com/google/go-github/github"
)

var client *github.Client
var ex *exit.Exit

func init() {
	ex = exit.Default()
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run seec.go <sha1>")
		ex.Exit(0)
	}
	sha1 := os.Args[1]
	client = github.NewClient(nil)
	displayRateLimit()
	commit, _, err := client.Git.GetCommit("git", "git", sha1)
	if err != nil {
		fmt.Printf("Unable to get commit '%s': err '%v'\n", sha1, err)
		ex.Exit(1)
	}
	if len(commit.Parents) != 2 {
		fmt.Printf("Sha1 '%s' has '%d' parent(s) instead of 2\n", sha1, len(commit.Parents))
	}
	clogin := login(*commit.Author.Email, *commit.Author.Name)
	parent := commit.Parents[1]
	var pcommit *github.Commit
	for pcommit == nil {
		pcommit, _, err = client.Git.GetCommit("git", "git", *parent.SHA)
		if err != nil {
			fmt.Printf("Unable to get parent commit '%s': err '%v'\n", parent.SHA, err)
			ex.Exit(1)
		}
		// fmt.Printf("pcommit '%+v', len %d\n", pcommit, len(pcommit.Parents))
		if len(pcommit.Parents) == 2 {
			parent = pcommit.Parents[1]
			pcommit = nil
		} else {
			break
		}
	}
	plogin := login(*pcommit.Author.Email, *pcommit.Author.Name)

	res := fmt.Sprintf("See [commit %s](https://github.com/git/git/commit/%s) by [%s (`%s`)](https://github.com/%s), %s.  \n",
		(*pcommit.SHA)[:7], *pcommit.SHA,
		*pcommit.Author.Name, plogin, plogin, pcommit.Author.Date.Format("02 Jan 2006"))
	res = res + fmt.Sprintf("<sup>(Merged by [%s -- `%s` --](https://github.com/%s) in [commit %s](https://github.com/git/git/commit/%s), %s)</sup>  ",
		*commit.Author.Name, clogin, clogin,
		sha1[:7], sha1, commit.Committer.Date.Format("02 Jan 2006"))
	res = collect(res, *pcommit.Message, "Test-adapted-from")
	res = collect(res, *pcommit.Message, "Helped-by")
	fmt.Println(res)
	clipboard.WriteAll(res)
	fmt.Println("(Copied to the clipboard)")
	displayRateLimit()
}

func displayRateLimit() {
	rate, _, err := client.RateLimit()
	if err != nil {
		fmt.Printf("Error fetching rate limit: %#v\n\n", err)
	} else {
		const layout = "15:04pm (MST)"
		t := rate.Reset.Time
		ts := fmt.Sprintf("%s", t.Format(layout))
		fmt.Printf("API Rate Limit: %d/%d (reset at %s)\n\n", rate.Remaining, rate.Limit, ts)
	}
}

func login(email string, name string) string {
	opts := &github.SearchOptions{Order: "desc"}
	var res *github.UsersSearchResult
	var err error
	if email != "" {
		res, _, err = client.Search.Users(email, opts)
		if err != nil {
			fmt.Printf("Unable to search user '%s': err '%v'", email, err)
			ex.Exit(1)
		}
	}
	if res == nil || *res.Total == 0 {
		name = strings.Replace(name, "-", " ", -1)
		res, _, err = client.Search.Users(name, opts)
		if err != nil {
			fmt.Printf("Unable to search user '%s': err '%v'", name, err)
			ex.Exit(1)
		}
	}
	if res == nil || *res.Total == 0 {
		var resIssues *github.IssuesSearchResult
		issueSearch := fmt.Sprintf(`"Signed-off-by: %s <%s>"`, name, email)
		resIssues, _, err = client.Search.Issues(issueSearch, opts)
		if err != nil {
			fmt.Printf("Unable to search issue '%s': err '%v'", issueSearch, err)
			ex.Exit(1)
		}
		if resIssues == nil || *resIssues.Total == 0 {
			return ""
		}
		issue := resIssues.Issues[0]
		return *issue.User.Login
	}
	if res == nil || *res.Total == 0 {
		return ""
	}
	return *res.Users[0].Login
}

func collect(res, msg, activity string) string {
	re := regexp.MustCompile(fmt.Sprintf(`%s:\s+([^<\r\n]+)\s+<([^>\r\n]+)>`, activity))
	activitymsg := activity + ": "
	first := true
	for _, resc := range re.FindAllStringSubmatch(msg, -1) {
		if len(resc) != 3 {
			continue
		}
		name := resc[1]
		email := resc[2]
		login := login(email, name)
		if !first {
			activitymsg = activitymsg + ", "
		}
		if login == "" {
			activitymsg = activitymsg + fmt.Sprintf("%s <%s>", name, email)
			first = false
			continue
		}
		activitymsg = activitymsg + fmt.Sprintf("[%s (`%s`)](https://github.com/%s)", name, login, login)
		first = false
	}
	if !first {
		res = res + "\n" + activitymsg + "  "
	}
	return res
}
