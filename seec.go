package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"seec2/internal/commits"
	"seec2/internal/gh"
	"seec2/internal/users"

	"github.com/VonC/godbg"
	"github.com/VonC/godbg/exit"
	"github.com/atotto/clipboard"
)

var ex *exit.Exit
var pdbg *godbg.Pdbg

func init() {
	ex = exit.Default()
	if os.Getenv("dbg") != "" {
		pdbg = godbg.NewPdbg()
	} else {
		pdbg = godbg.NewPdbg(godbg.OptExcludes([]string{"/seec.go"}))
	}
	gh.GHex = ex
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage:       go run seec.go <sha1>")
		fmt.Println("       dbg=1 go run seec.go <sha1> for debug information")
		fmt.Println(`       cmd /v /c "set dbg=1 && bin\seec* <sha1>" for debug information`)
		ex.Exit(0)
	}
	sha1 := os.Args[1]
	gh.DisplayRateLimit()
	commit := gh.MustGetCommit(sha1)
	pdbg.Pdbgf("commit: %s", commit)

	if commit.NbParents() > 2 {
		fmt.Printf("Sha1 '%s' has '%d' parent(s) instead of 2\n", sha1, commit.NbParents())
		ex.Exit(1)
	}

	date := commit.CommitterDate()

	parent := commit
	if commit.NbParents() == 2 {
		parent = commit.SecondParent()
	}
	var cuser *users.User
	if parent.SameSHA1(commit) {
		cuser = users.AuthorNameAndLogin(commit)
	} else {
		cuser = users.CommitterNameAndLogin(commit)
	}
	res := ""
	res = res + seeCommit(parent, commit)
	res = res + fmt.Sprintf("<sup>(Merged by [%s -- `%s` --](https://github.com/%s) in [commit %s](https://github.com/git/git/commit/%s), %s)</sup>  ",
		cuser.Name, cuser.Login, cuser.Login,
		sha1[:7], sha1, date)
	fmt.Println(res)
	clipboard.WriteAll(res)
	fmt.Println("(Copied to the clipboard)")
	gh.DisplayRateLimit()
}

func seeCommit(parent, commit *gh.Commit) string {
	pcommit := gh.FirstSingleParentCommit(parent)
	pdbg.Pdbgf("seeCommit '%s' => pcommit '%s'", parent, pcommit)
	apcommit := pcommit
	if pcommit.SameAuthor(commit) {
		pdbg.Pdbgf("Same author '%s', so call checkParentCommits\nInitial message: '%s'", *pcommit.Author.Name, *commit.Message)
		apcommit = pcommit.FirstParent()
	}
	pcommits := checkParentCommits(apcommit, *commit.Message)
	pdbg.Pdbgf("commitsByAuthors '%s'", pcommits)
	if len(pcommits) == 0 {
		pauthorname := pcommit.AuthorName()
		pcommitsByAuthor := commits.NewCommitsByAuthor(pauthorname)
		pcommitsByAuthor.AddCommit(pcommit)
		userAuthor := users.AuthorNameAndLogin(pcommit)
		pcommits[*userAuthor] = pcommitsByAuthor
		pdbg.Pdbgf("Put single commit '%s' for author '%s'", pcommitsByAuthor, pauthorname)
	}
	res := ""
	for userAuthor, pcommitsByAuthor := range pcommits {
		commitsbd := pcommitsByAuthor.CommitsByDate()
		first := true
		plogin := userAuthor.Login
		authorname := userAuthor.Name
		for i, cbd := range commitsbd {
			if first {
				res = res + "See "
			} else {
				res = res + ", "
			}
			first = false
			if i == len(commitsbd)-1 && i > 0 {
				res = res + "and "
			}
			commits := cbd.Commits()
			firstc := true
			for _, commit := range commits {
				if !firstc {
					res = res + ", "
				}
				firstc = false
				c := fmt.Sprintf("[commit %s](https://github.com/git/git/commit/%s)",
					(*commit.SHA)[:7], *commit.SHA)
				res = res + c
			}
			res = res + fmt.Sprintf(" (%s)", cbd.Date())
		}
		res = res + fmt.Sprintf(" by [%s (`%s`)](https://github.com/%s).  \n",
			authorname, plogin, plogin)
		// seec 8cc88166c00e555f1bf5375017ed91b7e2cc904e, https://github.com/git/git/commit/8cc88166c00e555f1bf5375017ed91b7e2cc904e
		res = collect(res, *pcommit.Message, "Suggested-by")
		// seec 777e75b60568b613e452ebbb30a1fb27c4fd7d8a, https://github.com/git/git/commit/777e75b60568b613e452ebbb30a1fb27c4fd7d8a
		res = collect(res, *pcommit.Message, "Test-adapted-from")
		// seec 6dec263333417738528089834bd8cda72017aa31, https://github.com/git/git/commit/6dec263333417738528089834bd8cda72017aa31
		// seec 324a9f41cbf96ad994efc3b20be239116eba0dae, https://github.com/git/git/commit/324a9f41cbf96ad994efc3b20be239116eba0dae
		res = collect(res, *pcommit.Message, "Helped-by")
	}
	return res
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
		login := users.Login(email, name)
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

func checkParentCommits(pcommit *gh.Commit, commitmsg string) commits.CommitsByAuthors {
	res := make(commits.CommitsByAuthors)
	pdbg.Pdbgf("pcommit '%s'", pcommit)
	pcommitmsgs := strings.Split(pcommit.MessageC(), "\n")
	pdbg.Pdbgf("pcommit message '%d'", len(pcommitmsgs))
	title := pcommitmsgs[0]
	pdbg.Pdbgf("title '%s'", title)
	if strings.Contains(commitmsg, title) {
		pauthorUser := *users.AuthorNameAndLogin(pcommit)
		pdbg.Pdbgf("pauthorname='%s' for '%s'", pauthorUser.Name, pauthorUser.Login)
		pcommitsByAuthor := res[pauthorUser]
		if pcommitsByAuthor == nil {
			pcommitsByAuthor = commits.NewCommitsByAuthor(pauthorUser.Name)
		}
		pcommitsByAuthor.AddCommit(pcommit)
		pdbg.Pdbgf("pcommitsByAuthor BEFOR='%s'", pcommitsByAuthor)
		res[pauthorUser] = pcommitsByAuthor
		pdbg.Pdbgf("res BEFOR='%s'", res)
		pdbg.Pdbgf("call checkParentCommits with parents '%+v', pca '%s' for '%s'",
			pcommit.Parents, pcommitsByAuthor.String(), pauthorUser.Name)
		ppcommits := checkParentCommits(pcommit.FirstParent(), commitmsg)
		res.Add(ppcommits)
		pdbg.Pdbgf("res AFTER='%s'", res)
	}
	return res
}
