package main

import (
	"fmt"
	"os"
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
	commits := checkParentCommits(apcommit, *commit.Message)
	pdbg.Pdbgf("commitsByAuthors '%s'", commits)
	return ""
}

func checkParentCommits(pcommit *gh.Commit, commitmsg string) commits.CommitsByAuthors {
	res := make(commits.CommitsByAuthors)
	pdbg.Pdbgf("pcommit '%s'", pcommit)
	pcommitmsgs := strings.Split(pcommit.MessageC(), "\n")
	pdbg.Pdbgf("pcommit message '%d'", len(pcommitmsgs))
	title := pcommitmsgs[0]
	pdbg.Pdbgf("title '%s'", title)
	if strings.Contains(commitmsg, title) {
		pauthorname := pcommit.AuthorName()
		pdbg.Pdbgf("pauthorname='%s' for '%v'", pauthorname, pcommit.Author)
		pcommitsByAuthor := res[pauthorname]
		if pcommitsByAuthor == nil {
			pcommitsByAuthor = commits.NewCommitsByAuthor(pauthorname)
		}
		pcommitsByAuthor.AddCommit(pcommit)
		pdbg.Pdbgf("pcommitsByAuthor BEFOR='%s'", pcommitsByAuthor)
		res[pauthorname] = pcommitsByAuthor
		pdbg.Pdbgf("res BEFOR='%s'", res)
		pdbg.Pdbgf("call checkParentCommits with parents '%+v', pca '%s' for '%s'",
			pcommit.Parents, pcommitsByAuthor.String(), pauthorname)
		ppcommits := checkParentCommits(pcommit.FirstParent(), commitmsg)
		res.Add(ppcommits)
		pdbg.Pdbgf("res AFTER='%s'", res)
	}
	return res
}
