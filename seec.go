package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
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
		fmt.Println(`       cmd /v /c "set dbg=1 && bin\seec* <sha1>" for debug information`)
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
	res := ""
	var parent github.Commit
	if len(commit.Parents) == 2 {
		parent = commit.Parents[1]
	}
	if len(commit.Parents) > 2 {
		fmt.Printf("Sha1 '%s' has '%d' parent(s) instead of 2\n", sha1, len(commit.Parents))
		ex.Exit(1)
	}
	// fmt.Printf("commit='%+v'\n", commit)
	// fmt.Printf("commit.Author='%+v'\n", commit.Author)
	clogin := ""
	cname := ""
	if *parent.SHA != *commit.SHA {
		clogin = login(*commit.Author.Email, *commit.Author.Name, *commit.SHA)
		cname = *commit.Author.Name
		res = res + seeCommit(&parent, commit)
	} else {
		clogin = login(*commit.Committer.Email, *commit.Committer.Name, "")
		cname = *commit.Committer.Name
	}
	// fmt.Printf("clogin='%s'", clogin)
	// ex.Exit(0)
	res = res + fmt.Sprintf("<sup>(Merged by [%s -- `%s` --](https://github.com/%s) in [commit %s](https://github.com/git/git/commit/%s), %s)</sup>  ",
		cname, clogin, clogin,
		sha1[:7], sha1, commit.Committer.Date.Format("02 Jan 2006"))
	fmt.Println(res)
	clipboard.WriteAll(res)
	fmt.Println("(Copied to the clipboard)")
	displayRateLimit()
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
	pdbg.Pdbgf("Parent commit '%s'", *pcommit.SHA)
	return "_"
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

func login(email string, name string, SHA string) string {
	fmt.Printf("email='%s', name='%s'\n", email, name)
	if login := cacheLogins(email, name); login != "" {
		return login
	}
	if login := scrapPage(SHA); login != "" {
		return addToCacheLogins(email, name, login)
	}
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
	nameNoDash := name
	if res == nil || *res.Total == 0 {
		nameNoDash = strings.Replace(name, "-", " ", -1)
		res, _, err = client.Search.Users(nameNoDash, opts)
		if err != nil {
			fmt.Printf("Unable to search user '%s': err '%v'", nameNoDash, err)
			ex.Exit(1)
		}
	}
	if res == nil || *res.Total == 0 {
		var resIssues *github.IssuesSearchResult
		issueSearch := fmt.Sprintf(`"Signed-off-by: %s <%s>"`, nameNoDash, email)
		resIssues, _, err = client.Search.Issues(issueSearch, opts)
		if err != nil {
			fmt.Printf("Unable to search issue '%s': err '%v'", issueSearch, err)
			ex.Exit(1)
		}
		if resIssues == nil || *resIssues.Total == 0 {
			return ""
		}
		issue := resIssues.Issues[0]
		return addToCacheLogins(email, name, *issue.User.Login)
	}
	if res == nil || *res.Total == 0 {
		return ""
	}
	return addToCacheLogins(email, name, *res.Users[0].Login)
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
		login := login(email, name, "")
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

var cacheLoginsRE = regexp.MustCompile(`^(.*)#~#(.*?)\^\((.*?)\)\^(.*?)$`)

func cacheLogins(email string, name string) string {
	login := ""
	fi, err := os.OpenFile("logins.txt", os.O_RDONLY|os.O_CREATE, 0660)
	if err != nil {
		panic(err)
	}
	// close fi on exit and check for its returned error
	defer func() {
		if err := fi.Close(); err != nil {
			panic(err)
		}
	}()
	var part []byte
	var prefix bool
	reader := bufio.NewReader(fi)
	buffer := bytes.NewBuffer(make([]byte, 0))
	for {
		if part, prefix, err = reader.ReadLine(); err != nil {
			break
		}
		buffer.Write(part)
		if !prefix {
			line := buffer.String()
			re := cacheLoginsRE.FindAllStringSubmatch(line, -1)
			if len(re) == 1 {
				femail := re[0][1]
				fname := re[0][2]
				flogin := re[0][4]
				if femail == email && fname == name {
					// fmt.Printf("femail='%s', fname='%s' => login '%s'\n", femail, fname, flogin)
					login = flogin
					break
				}
			}
			buffer.Reset()
		}
	}
	if err == io.EOF {
		err = nil
	}
	if err != nil {
		panic(err)
	}
	fmt.Printf("cacheLogins='%s'\n", login)
	return login
}

var contributorRE = regexp.MustCompile(`.*?<a\s+href="[^"]+"\s+rel="contributor"\s*?>([^<]+)</a>`)

func scrapPage(SHA string) string {
	if SHA == "" {
		return ""
	}
	response, err := http.Get("https://github.com/git/git/commit/" + SHA)
	if err != nil {
		fmt.Printf("%s", err)
		os.Exit(1)
	} else {
		defer response.Body.Close()
		contents, err := ioutil.ReadAll(response.Body)
		if err != nil {
			fmt.Printf("%s", err)
			os.Exit(1)
		}
		lines := strings.Split(string(contents), `\n`)
		for _, line := range lines {
			if strings.Contains(line, "contributor") {
				r := contributorRE.FindAllStringSubmatch(line, -1)
				if len(r) > 0 {
					if len(r[0]) == 2 {
						// fmt.Printf("r[0]='%+v' => res='%s'\n", r[0], r[0][1])
						return r[0][1]
					}
				}
				fmt.Println("line='%+v'\n", r)
			}
		}
	}
	return ""
}
func addToCacheLogins(email string, name string, login string) string {
	fmt.Printf("addToCacheLogins '%s'\n", login)
	fi, err := os.OpenFile("logins.txt", os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)
	if err != nil {
		panic(err)
	}
	// close fi on exit and check for its returned error
	defer func() {
		if err := fi.Close(); err != nil {
			panic(err)
		}
	}()
	if _, err := fi.WriteString(fmt.Sprintf("%s#~#%s^()^%s\n", email, name, login)); err != nil {
		panic(err)
	}
	return login

}
