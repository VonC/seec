package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/VonC/godbg"
	"github.com/VonC/godbg/exit"
	"github.com/atotto/clipboard"
	"github.com/google/go-github/github"
)

var client *github.Client
var ex *exit.Exit
var pdbg *godbg.Pdbg

func init() {
	ex = exit.Default()
	if os.Getenv("dbg") != "" {
		pdbg = godbg.NewPdbg()
	} else {
		pdbg = godbg.NewPdbg(godbg.OptExcludes([]string{"/seec.go"}))
	}
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage:       go run seec.go <sha1>")
		fmt.Println("       dbg=1 go run seec.go <sha1> for debug information")
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
	res := ""
	res = res + seeCommit(&parent, commit)
	res = res + fmt.Sprintf("<sup>(Merged by [%s -- `%s` --](https://github.com/%s) in [commit %s](https://github.com/git/git/commit/%s), %s)</sup>  ",
		*commit.Author.Name, clogin, clogin,
		sha1[:7], sha1, commit.Committer.Date.Format("02 Jan 2006"))
	fmt.Println(res)
	clipboard.WriteAll(res)
	fmt.Println("(Copied to the clipboard)")
	displayRateLimit()
}

func displayRateLimit() {
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

type commitsByAuthor struct {
	author *github.CommitAuthor
	cbd    []*commitsByDate
}

// Because of seec 709cd912d4663af87903d3d278a3bab9d4d84153
type commitsByDate struct {
	date    *time.Time
	commits []*github.Commit
}

func (cba *commitsByAuthor) String() string {
	res := ""
	first := true
	for i, acbd := range cba.cbd {
		if !first {
			res = res + ", "
		}
		first = false
		if i == len(cba.cbd)-1 && i > 0 {
			res = res + "and "
		}
		res = res + acbd.String()
	}
	return fmt.Sprintf("%s=>%s", *cba.author.Name, res)
}

func (cbd *commitsByDate) String() string {
	res := ""
	first := true
	for i, commit := range cbd.commits {
		if !first {
			res = res + ", "
		}
		first = false
		if i == len(cbd.commits)-1 && i > 0 {
			res = res + "and "
		}
		res = res + (*commit.SHA)[:7]
	}
	return fmt.Sprintf("%s (%s)", res, cbd.date.Format("02 Jan 2006"))
}

func seeCommit(parent, commit *github.Commit) string {
	var pcommit *github.Commit
	var err error
	for pcommit == nil {
		pcommit, _, err = client.Git.GetCommit("git", "git", *parent.SHA)
		if err != nil {
			fmt.Printf("Unable to get parent commit '%s': err '%v'\n", parent.SHA, err)
			ex.Exit(1)
		}
		// fmt.Printf("pcommit '%+v', len %d\n", pcommit, len(pcommit.Parents))
		if len(pcommit.Parents) == 2 {
			parent = &pcommit.Parents[1]
			pcommit = nil
		} else {
			break
		}
	}
	var commits = make(map[string]*commitsByAuthor)
	apcommit := pcommit
	if *pcommit.Author.Name == *commit.Author.Name {
		pdbg.Pdbgf("Same author '%s', so call checkParentCommits\nInitial message: '%s'", *pcommit.Author.Name, *commit.Message)
		apcommit = &pcommit.Parents[0]
	}
	commits = checkParentCommits(apcommit, *commit.Message)
	if len(commits) == 0 {
		pauthorname := *pcommit.Author.Name
		pcommitsByAuthor := &commitsByAuthor{pcommit.Author, []*github.Commit{pcommit}}
		commits[pauthorname] = pcommitsByAuthor
		pdbg.Pdbgf("Put single commit '%s' for author '%s'", pcommitsByAuthor, pauthorname)
	}
	res := ""
	for _, pcommitsByAuthor := range commits {
		author := pcommitsByAuthor.author
		pcommits := pcommitsByAuthor.pcommits
		plogin := login(*author.Email, *author.Name)
		first := true
		for i, pcommit := range pcommits {
			if first {
				res = "See "
			} else {
				res = res + ", "
			}
			first = false
			if i == len(pcommits)-1 && i > 0 {
				res = res + "and "
			}
			c := fmt.Sprintf("[commit %s](https://github.com/git/git/commit/%s) [%s]",
				(*pcommit.SHA)[:7], *pcommit.SHA, pcommit.Author.Date.Format("02 Jan 2006"))
			res = res + c
		}
		res = res + fmt.Sprintf(" by [%s (`%s`)](https://github.com/%s).  \n",
			*author.Name, plogin, plogin)
		// seec 777e75b60568b613e452ebbb30a1fb27c4fd7d8a, https://github.com/git/git/commit/777e75b60568b613e452ebbb30a1fb27c4fd7d8a
		res = collect(res, *pcommit.Message, "Test-adapted-from")
		// seec 6dec263333417738528089834bd8cda72017aa31, https://github.com/git/git/commit/6dec263333417738528089834bd8cda72017aa31
		// seec 324a9f41cbf96ad994efc3b20be239116eba0dae, https://github.com/git/git/commit/324a9f41cbf96ad994efc3b20be239116eba0dae
		res = collect(res, *pcommit.Message, "Helped-by")
	}
	return res
}

// for cases like commit a6be52e239df4d4a469a5324273f43a0695fe95d
func checkParentCommits(apcommit *github.Commit, commitmsg string) map[string]*commitsByAuthor {
	res := make(map[string]*commitsByAuthor)
	pcommit, _, err := client.Git.GetCommit("git", "git", *apcommit.SHA)
	if err != nil {
		fmt.Printf("Unable to get check parent commit '%s': err '%v'\n", *apcommit.SHA, err)
		ex.Exit(1)
	}
	pdbg.Pdbgf("pcommit %s", *pcommit.SHA)
	pcommitmsgs := strings.Split(*pcommit.Message, "\n")
	title := pcommitmsgs[0]
	pdbg.Pdbgf("title '%s'", title)
	if strings.Contains(commitmsg, title) {
		pauthorname := *pcommit.Author.Name
		pdbg.Pdbgf("pauthorname='%s' for '%v'", pauthorname, pcommit.Author)
		pcommitsByAuthor := res[pauthorname]
		if pcommitsByAuthor == nil {
			pcommitsByAuthor = &commitsByAuthor{pcommit.Author, []*github.Commit{}}
		}
		pcommitsByAuthor.pcommits = append(pcommitsByAuthor.pcommits, pcommit)
		res[pauthorname] = pcommitsByAuthor
		pdbg.Pdbgf("call checkParentCommits with parents '%+v', pca '%s' for '%s'",
			pcommit.Parents, pcommitsByAuthor.String(), pauthorname)
		ppcommits := checkParentCommits(&pcommit.Parents[0], commitmsg)
		for authorName, pcommitsByAuthor := range ppcommits {
			acommitsByAuthor := res[authorName]
			if acommitsByAuthor == nil {
				res[authorName] = pcommitsByAuthor
			} else {
				for _, pc := range pcommitsByAuthor.pcommits {
					acommitsByAuthor.pcommits = append(acommitsByAuthor.pcommits, pc)
				}
				res[authorName] = acommitsByAuthor
				pdbg.Pdbgf("Put commits '%s' for author '%s'", acommitsByAuthor.String(), authorName)
			}
		}
	}
	return res
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
	allresc := re.FindAllStringSubmatch(msg, -1)
	for i, resc := range allresc {
		dot := ""
		if len(resc) != 3 {
			continue
		}
		name := resc[1]
		email := resc[2]
		login := login(email, name)
		if !first {
			activitymsg = activitymsg + ", "
		}
		if i == len(allresc)-1 {
			dot = "."
			if i > 0 {
				activitymsg = activitymsg + "and "
			}
		}
		if login == "" {
			activitymsg = activitymsg + fmt.Sprintf("%s <%s>%s", name, email, dot)
			first = false
			continue
		}
		activitymsg = activitymsg + fmt.Sprintf("[%s (`%s`)](https://github.com/%s)%s", name, login, login, dot)
		first = false
	}
	if !first {
		res = res + activitymsg + "  \n"
	}
	return res
}
